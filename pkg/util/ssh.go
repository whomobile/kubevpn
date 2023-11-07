package util

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kevinburke/ssh_config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/util/homedir"

	"github.com/wencaiwulue/kubevpn/pkg/daemon/rpc"
	"github.com/wencaiwulue/kubevpn/pkg/errors"
)

type SshConfig struct {
	Addr             string
	User             string
	Password         string
	Keyfile          string
	ConfigAlias      string
	RemoteKubeconfig string
}

func ParseSshFromRPC(sshJump *rpc.SshJump) *SshConfig {
	if sshJump == nil {
		return &SshConfig{}
	}
	return &SshConfig{
		Addr:             sshJump.Addr,
		User:             sshJump.User,
		Password:         sshJump.Password,
		Keyfile:          sshJump.Keyfile,
		ConfigAlias:      sshJump.ConfigAlias,
		RemoteKubeconfig: sshJump.RemoteKubeconfig,
	}
}

func (s *SshConfig) ToRPC() *rpc.SshJump {
	return &rpc.SshJump{
		Addr:             s.Addr,
		User:             s.User,
		Password:         s.Password,
		Keyfile:          s.Keyfile,
		ConfigAlias:      s.ConfigAlias,
		RemoteKubeconfig: s.RemoteKubeconfig,
	}
}

func Main(pctx context.Context, remoteEndpoint, localEndpoint netip.AddrPort, conf *SshConfig, done chan struct{}) error {
	ctx, cancelFunc := context.WithCancel(pctx)
	defer cancelFunc()

	sshClient, err := DialSshRemote(conf)
	if err != nil {
		errors.LogErrorf("Dial into remote server error: %s", err)
		return err
	}

	// ref: https://github.com/golang/go/issues/21478
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		select {
		case <-ticker.C:
			_, _, err := sshClient.SendRequest("keepalive@golang.org", true, nil)
			if err != nil {
				errors.LogErrorf("failed to send keep alive error: %s", err)
			}
		case <-ctx.Done():
			return
		}
	}()

	// Listen on remote server port
	var lc net.ListenConfig
	listen, err := lc.Listen(ctx, "tcp", localEndpoint.String())
	if err != nil {
		err = errors.Wrap(err, "lc.Listen(ctx, \"tcp\", localEndpoint.String()): ")
		return err
	}
	defer listen.Close()

	select {
	case done <- struct{}{}:
	default:
	}
	// handle incoming connections on reverse forwarded tunnel
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		local, err := listen.Accept()
		if err != nil {
			err = errors.Wrap(err, "listen.Accept(): ")
			return err
		}
		go func(local net.Conn) {
			defer local.Close()
			conn, err := sshClient.Dial("tcp", remoteEndpoint.String())
			if err != nil {
				errors.LogErrorf("Failed to dial %s: %s", remoteEndpoint.String(), err)
				cancelFunc()
				return
			}
			defer conn.Close()
			handleClient(local, conn)
		}(local)
	}
}

// todo ssh heartbeats
// https://github.com/golang/go/issues/21478
func DialSshRemote(conf *SshConfig) (*ssh.Client, error) {
	var remote *ssh.Client
	var err error
	if conf.ConfigAlias != "" {
		remote, err = jumpRecursion(conf.ConfigAlias)
	} else {
		var auth []ssh.AuthMethod
		if conf.Password != "" {
			auth = append(auth, ssh.Password(conf.Password))
		} else {
			if conf.Keyfile == "" {
				conf.Keyfile = filepath.Join(homedir.HomeDir(), ".ssh", "id_rsa")
			}
			var keyFile ssh.AuthMethod
			keyFile, err = publicKeyFile(conf.Keyfile)
			if err != nil {
				err = errors.Wrap(err, "publicKeyFile(conf.Keyfile): ")
				return nil, err
			}
			auth = append(auth, keyFile)
		}

		// refer to https://godoc.org/golang.org/x/crypto/ssh for other authentication types
		sshConfig := &ssh.ClientConfig{
			// SSH connection username
			User:            conf.User,
			Auth:            auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			BannerCallback:  ssh.BannerDisplayStderr(),
			Timeout:         time.Second * 10,
		}
		if strings.Index(conf.Addr, ":") < 0 {
			// use default ssh port 22
			conf.Addr = net.JoinHostPort(conf.Addr, "22")
		}
		// Connect to SSH remote server using serverEndpoint
		remote, err = ssh.Dial("tcp", conf.Addr, sshConfig)
	}
	return remote, err
}

func RemoteRun(conf *SshConfig, cmd string, env map[string]string) (output []byte, errOut []byte, err error) {
	var remote *ssh.Client
	remote, err = DialSshRemote(conf)
	if err != nil {
		errors.LogErrorf("Dial into remote server error: %s", err)
		return
	}
	defer remote.Close()
	var session *ssh.Session
	session, err = remote.NewSession()
	if err != nil {
		err = errors.Wrap(err, "remote.NewSession(): ")
		return
	}
	for k, v := range env {
		// /etc/ssh/sshd_config
		// AcceptEnv DEBIAN_FRONTEND
		if err = session.Setenv(k, v); err != nil {
			log.Warn(err)
			err = nil
		}
	}
	defer remote.Close()
	var out bytes.Buffer
	var er bytes.Buffer
	session.Stdout = &out
	session.Stderr = &er
	err = session.Run(cmd)
	return out.Bytes(), er.Bytes(), err
}

func publicKeyFile(file string) (ssh.AuthMethod, error) {
	var err error
	if len(file) != 0 && file[0] == '~' {
		file = filepath.Join(homedir.HomeDir(), file[1:])
	}
	file, err = filepath.Abs(file)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("Cannot read SSH public key file %s", file))
		return nil, err
	}
	buffer, err := os.ReadFile(file)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("Cannot read SSH public key file %s", file))
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("Cannot parse SSH public key file %s", file))
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func handleClient(client net.Conn, remote net.Conn) {
	chDone := make(chan bool, 2)

	// start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			log.Debugf("error while copy remote->local: %s", err)
		}
		select {
		case chDone <- true:
		default:
		}
	}()

	// start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			log.Debugf("error while copy local->remote: %s", err)
		}
		select {
		case chDone <- true:
		default:
		}
	}()

	<-chDone
}

func jumpRecursion(name string) (client *ssh.Client, err error) {
	var jumper = "ProxyJump"
	var bastionList = []*SshConfig{getBastion(name)}
	for {
		value := confList.Get(name, jumper)
		if value != "" {
			bastionList = append(bastionList, getBastion(value))
			name = value
			continue
		}
		break
	}
	for i := len(bastionList) - 1; i >= 0; i-- {
		if bastionList[i] == nil {
			return nil, errors.New("config is nil")
		}
		if client == nil {
			client, err = dial(bastionList[i])
			if err != nil {
				err = errors.Wrap(err, "dial(bastionList[i]): ")
				return
			}
		} else {
			client, err = jump(client, bastionList[i])
			if err != nil {
				err = errors.Wrap(err, "jump(client, bastionList[i]): ")
				return
			}
		}
	}
	return
}

func getBastion(name string) *SshConfig {
	var host, port string
	config := SshConfig{
		ConfigAlias: name,
	}
	var propertyList = []string{"ProxyJump", "Hostname", "User", "Port", "IdentityFile"}
	for i, s := range propertyList {
		value := confList.Get(name, s)
		switch i {
		case 0:

		case 1:
			host = value
		case 2:
			config.User = value
		case 3:
			if port = value; port == "" {
				port = strconv.Itoa(22)
			}
		case 4:
			config.Keyfile = value
		}
	}
	config.Addr = net.JoinHostPort(host, port)
	return &config
}

func dial(from *SshConfig) (*ssh.Client, error) {
	// connect to the bastion host
	authMethod, err := publicKeyFile(from.Keyfile)
	if err != nil {
		err = errors.Wrap(err, "publicKeyFile(from.Keyfile): ")
		return nil, err
	}
	return ssh.Dial("tcp", from.Addr, &ssh.ClientConfig{
		User:            from.User,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback:  ssh.BannerDisplayStderr(),
		Timeout:         time.Second * 10,
	})
}

func jump(bClient *ssh.Client, to *SshConfig) (*ssh.Client, error) {
	// Dial a connection to the service host, from the bastion
	conn, err := bClient.Dial("tcp", to.Addr)
	if err != nil {
		err = errors.Wrap(err, "bClient.Dial(\"tcp\", to.Addr): ")
		return nil, err
	}

	authMethod, err := publicKeyFile(to.Keyfile)
	if err != nil {
		err = errors.Wrap(err, "publicKeyFile(to.Keyfile): ")
		return nil, err
	}
	ncc, chans, reqs, err := ssh.NewClientConn(conn, to.Addr, &ssh.ClientConfig{
		User:            to.User,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback:  ssh.BannerDisplayStderr(),
		Timeout:         time.Second * 10,
	})
	if err != nil {
		return nil, err
	}

	sClient := ssh.NewClient(ncc, chans, reqs)
	return sClient, nil
}

type conf []*ssh_config.Config

func (c conf) Get(alias string, key string) string {
	for _, s := range c {
		if v, err := s.Get(alias, key); err == nil {
			return v
		}
	}
	return ssh_config.Get(alias, key)
}

var once sync.Once
var confList conf

func init() {
	once.Do(func() {
		strings := []string{
			filepath.Join(homedir.HomeDir(), ".ssh", "config"),
			filepath.Join("/", "etc", "ssh", "ssh_config"),
		}
		for _, s := range strings {
			file, err := os.ReadFile(s)
			if err != nil {
				continue
			}
			cfg, err := ssh_config.DecodeBytes(file)
			if err != nil {
				continue
			}
			confList = append(confList, cfg)
		}
	})
}
