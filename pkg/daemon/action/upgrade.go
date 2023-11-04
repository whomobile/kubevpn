package action

import (
	"context"
	"errors"

	goversion "github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"

	"github.com/wencaiwulue/kubevpn/pkg/config"
	"github.com/wencaiwulue/kubevpn/pkg/daemon/rpc"
)

func (svr *Server) Upgrade(ctx context.Context, req *rpc.UpgradeRequest) (*rpc.UpgradeResponse, error) {
	var err error
	var clientVersion, daemonVersion *goversion.Version
	clientVersion, err = goversion.NewVersion(req.ClientVersion)
	if err != nil {
		err = errors.New("goversion.NewVersion(req.ClientVersion): " + err.Error())
		return nil, err
	}
	daemonVersion, err = goversion.NewVersion(config.Version)
	if err != nil {
		err = errors.New("goversion.NewVersion(config.Version): " + err.Error())
		return nil, err
	}
	if clientVersion.GreaterThan(daemonVersion) || (clientVersion.Equal(daemonVersion) && req.ClientCommitId != config.GitCommit) {
		log.Info("daemon version is less than client, needs to upgrade")
		return &rpc.UpgradeResponse{NeedUpgrade: true}, nil
	}
	return &rpc.UpgradeResponse{NeedUpgrade: false}, nil
}
