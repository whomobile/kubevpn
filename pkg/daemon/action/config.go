package action

import (
	"context"
	"errors"

	"github.com/spf13/pflag"
	"github.com/wencaiwulue/kubevpn/pkg/handler"
	"github.com/wencaiwulue/kubevpn/pkg/util"

	"github.com/wencaiwulue/kubevpn/pkg/daemon/rpc"
)

var CancelFunc = make(map[string]context.CancelFunc)

func (svr *Server) ConfigAdd(ctx context.Context, req *rpc.ConfigAddRequest) (*rpc.ConfigAddResponse, error) {
	var sshConf = util.ParseSshFromRPC(req.SshJump)
	file, err := util.ConvertToTempKubeconfigFile([]byte(req.KubeconfigBytes))
	if err != nil {
		err = errors.New("util.ConvertToTempKubeconfigFile([]byte(req.KubeconfigBytes)): " + err.Error())
		return nil, err
	}
	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.AddFlag(&pflag.Flag{
		Name:     "kubeconfig",
		DefValue: file,
	})
	sshCtx, sshCancel := context.WithCancel(context.Background())
	var path string
	path, err = handler.SshJump(sshCtx, sshConf, flags, true)
	CancelFunc[path] = sshCancel
	if err != nil {
		err = errors.New("sshCancel: " + err.Error())
		return nil, err
	}

	return &rpc.ConfigAddResponse{ClusterID: path}, nil
}

func (svr *Server) ConfigRemove(ctx context.Context, req *rpc.ConfigRemoveRequest) (*rpc.ConfigRemoveResponse, error) {
	if cancel, ok := CancelFunc[req.ClusterID]; ok {
		cancel()
	}
	return &rpc.ConfigRemoveResponse{}, nil
}
