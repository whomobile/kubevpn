package action

import (
	"context"
	"io"
	defaultlog "log"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/utils/pointer"

	"github.com/wencaiwulue/kubevpn/pkg/config"
	"github.com/wencaiwulue/kubevpn/pkg/daemon/rpc"
	"github.com/wencaiwulue/kubevpn/pkg/errors"
	"github.com/wencaiwulue/kubevpn/pkg/handler"
	"github.com/wencaiwulue/kubevpn/pkg/util"
)

func (svr *Server) Connect(req *rpc.ConnectRequest, resp rpc.Daemon_ConnectServer) error {
	defer func() {
		log.SetOutput(svr.LogFile)
		log.SetLevel(log.DebugLevel)
	}()
	out := io.MultiWriter(newWarp(resp), svr.LogFile)
	log.SetOutput(out)
	log.SetLevel(log.InfoLevel)
	if !svr.IsSudo {
		return svr.redirectToSudoDaemon(req, resp)
	}

	ctx := resp.Context()
	if !svr.t.IsZero() {
		log.Debugf("already connect to another cluster, you can disconnect this connect by command `kubevpn disconnect`")
		// todo define already connect error?
		return status.Error(codes.AlreadyExists, "already connect to another cluster, you can disconnect this connect by command `kubevpn disconnect`")
	}
	svr.t = time.Now()
	svr.connect = &handler.ConnectOptions{
		Namespace:            req.Namespace,
		Headers:              req.Headers,
		Workloads:            req.Workloads,
		ExtraCIDR:            req.ExtraCIDR,
		ExtraDomain:          req.ExtraDomain,
		UseLocalDNS:          req.UseLocalDNS,
		Engine:               config.Engine(req.Engine),
		OriginKubeconfigPath: req.OriginKubeconfigPath,
	}
	var sshConf = util.ParseSshFromRPC(req.SshJump)
	var transferImage = req.TransferImage

	go util.StartupPProf(config.PProfPort)
	defaultlog.Default().SetOutput(io.Discard)
	if transferImage {
		err := util.TransferImage(ctx, sshConf, config.OriginImage, req.Image, out)
		if err != nil {
			err = errors.Wrap(err, "util.TransferImage(ctx, sshConf, config.OriginImage, req.Image, out): ")
			return err
		}
	}
	file, err := util.ConvertToTempKubeconfigFile([]byte(req.KubeconfigBytes))
	if err != nil {
		err = errors.Wrap(err, "util.ConvertToTempKubeconfigFile([]byte(req.KubeconfigBytes)): ")
		return err
	}
	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.AddFlag(&pflag.Flag{
		Name:     "kubeconfig",
		DefValue: file,
	})

	sshCtx, sshCancel := context.WithCancel(context.Background())
	svr.connect.AddRolloutFunc(func() error {
		sshCancel()
		return nil
	})
	var path string
	path, err = handler.SshJump(sshCtx, sshConf, flags, false)
	if err != nil {
		err = errors.Wrap(err, "handler.SshJump(sshCtx, sshConf, flags, false): ")
		return err
	}
	err = svr.connect.InitClient(InitFactoryByPath(path, req.Namespace))
	if err != nil {
		err = errors.Wrap(err, "svr.connect.InitClient(InitFactoryByPath(path, req.Namespace)): ")
		return err
	}
	err = svr.connect.PreCheckResource()
	if err != nil {
		err = errors.Wrap(err, "svr.connect.PreCheckResource(): ")
		return err
	}
	_, err = svr.connect.RentInnerIP(ctx)
	if err != nil {
		err = errors.Wrap(err, "svr.connect.RentInnerIP(ctx): ")
		return err
	}

	config.Image = req.Image
	err = svr.connect.DoConnect(sshCtx, false)
	if err != nil {
		log.Errorf("do connect error: %v", err)
		svr.connect.Cleanup()
		svr.connect = nil
		svr.t = time.Time{}
		return err
	}
	return nil
}

func (svr *Server) redirectToSudoDaemon(req *rpc.ConnectRequest, resp rpc.Daemon_ConnectServer) error {
	cli := svr.GetClient(true)
	if cli == nil {
		return errors.Errorf("sudo daemon not start")
	}
	connect := &handler.ConnectOptions{
		Namespace:            req.Namespace,
		Headers:              req.Headers,
		Workloads:            req.Workloads,
		ExtraCIDR:            req.ExtraCIDR,
		ExtraDomain:          req.ExtraDomain,
		UseLocalDNS:          req.UseLocalDNS,
		Engine:               config.Engine(req.Engine),
		OriginKubeconfigPath: req.OriginKubeconfigPath,
	}
	var sshConf = util.ParseSshFromRPC(req.SshJump)
	file, err := util.ConvertToTempKubeconfigFile([]byte(req.KubeconfigBytes))
	if err != nil {
		err = errors.Wrap(err, "util.ConvertToTempKubeconfigFile([]byte(req.KubeconfigBytes)): ")
		return err
	}
	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.AddFlag(&pflag.Flag{
		Name:     "kubeconfig",
		DefValue: file,
	})
	sshCtx, sshCancel := context.WithCancel(context.Background())
	connect.AddRolloutFunc(func() error {
		sshCancel()
		return nil
	})
	var path string
	path, err = handler.SshJump(sshCtx, sshConf, flags, true)
	if err != nil {
		err = errors.Wrap(err, "handler.SshJump(sshCtx, sshConf, flags, true): ")
		return err
	}
	err = connect.InitClient(InitFactoryByPath(path, req.Namespace))
	if err != nil {
		err = errors.Wrap(err, "connect.InitClient(InitFactoryByPath(path, req.Namespace)): ")
		return err
	}
	err = connect.PreCheckResource()
	if err != nil {
		err = errors.Wrap(err, "connect.PreCheckResource(): ")
		return err
	}

	if svr.connect != nil {
		var isSameCluster bool
		isSameCluster, err = util.IsSameCluster(
			svr.connect.GetClientset().CoreV1().ConfigMaps(svr.connect.Namespace), svr.connect.Namespace,
			connect.GetClientset().CoreV1().ConfigMaps(connect.Namespace), connect.Namespace,
		)
		if err == nil && isSameCluster && svr.connect.Equal(connect) {
			// same cluster, do nothing
			log.Infof("already connect to cluster")
			return nil
		}
	}

	ctx, err := connect.RentInnerIP(resp.Context())
	if err != nil {
		err = errors.Wrap(err, "connect.RentInnerIP(resp.Context()): ")
		return err
	}

	connResp, err := cli.Connect(ctx, req)
	if err != nil {
		err = errors.Wrap(err, "cli.Connect(ctx, req): ")
		return err
	}
	for {
		recv, err := connResp.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		err = resp.Send(recv)
		if err != nil {
			err = errors.Wrap(err, "resp.Send(recv): ")
			return err
		}
	}

	svr.t = time.Now()
	svr.connect = connect

	// hangup
	if req.Foreground {
		<-resp.Context().Done()

		client := svr.GetClient(false)
		if client == nil {
			return errors.Errorf("daemon not start")
		}
		disconnect, err := client.Disconnect(context.Background(), &rpc.DisconnectRequest{
			ID: pointer.Int32(int32(0)),
		})
		if err != nil {
			log.Errorf("disconnect error: %v", err)
			return err
		}
		for {
			recv, err := disconnect.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Error(err)
				return err
			}
			log.Info(recv.Message)
		}
	}

	return nil
}

type warp struct {
	server rpc.Daemon_ConnectServer
}

func (r *warp) Write(p []byte) (n int, err error) {
	err = r.server.Send(&rpc.ConnectResponse{
		Message: string(p),
	})
	return len(p), err
}

func newWarp(server rpc.Daemon_ConnectServer) io.Writer {
	return &warp{server: server}
}

func InitFactory(kubeconfigBytes string, ns string) cmdutil.Factory {
	configFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	configFlags.WrapConfigFn = func(c *rest.Config) *rest.Config {
		if path, ok := os.LookupEnv(config.EnvSSHJump); ok {
			bytes, err := os.ReadFile(path)
			cmdutil.CheckErr(err)
			var conf *restclient.Config
			conf, err = clientcmd.RESTConfigFromKubeConfig(bytes)
			cmdutil.CheckErr(err)
			return conf
		}
		return c
	}
	// todo optimize here
	temp, err := os.CreateTemp("", "*.json")
	if err != nil {
		err = errors.Wrap(err, "os.CreateTemp(\"\", \"*.json\"): ")
		return nil
	}
	err = temp.Close()
	if err != nil {
		err = errors.Wrap(err, "temp.Close(): ")
		return nil
	}
	err = os.WriteFile(temp.Name(), []byte(kubeconfigBytes), os.ModePerm)
	if err != nil {
		err = errors.Wrap(err, "os.WriteFile(temp.Name(), []byte(kubeconfigBytes), os.ModePerm): ")
		return nil
	}
	configFlags.KubeConfig = pointer.String(temp.Name())
	configFlags.Namespace = pointer.String(ns)
	matchVersionFlags := cmdutil.NewMatchVersionFlags(configFlags)
	return cmdutil.NewFactory(matchVersionFlags)
}

func InitFactoryByPath(kubeconfig string, ns string) cmdutil.Factory {
	configFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	configFlags.KubeConfig = pointer.String(kubeconfig)
	configFlags.Namespace = pointer.String(ns)
	matchVersionFlags := cmdutil.NewMatchVersionFlags(configFlags)
	return cmdutil.NewFactory(matchVersionFlags)
}
