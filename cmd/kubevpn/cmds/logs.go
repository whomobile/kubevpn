package cmds

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/wencaiwulue/kubevpn/pkg/daemon"
	"github.com/wencaiwulue/kubevpn/pkg/daemon/rpc"
)

func CmdLogs(f cmdutil.Factory) *cobra.Command {
	req := &rpc.LogRequest{}
	cmd := &cobra.Command{
		Use:   "logs",
		Short: i18n.T("Log kubevpn daemon server"),
		Long:  templates.LongDesc(i18n.T(`Log kubevpn daemon server`)),
		Example: templates.Examples(i18n.T(`
        # show log for kubevpn daemon server
        kubevpn logs
        # follow more log
        kubevpn logs -f
`)),
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			// startup daemon process and sudo process
			return daemon.StartupDaemon(cmd.Context())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := daemon.GetClient(true).Logs(cmd.Context(), req)
			if err != nil {
				return err
			}
			var resp *rpc.LogResponse
			for {
				resp, err = client.Recv()
				if err == io.EOF {
					break
				} else if err == nil {
					fmt.Fprintln(os.Stdout, resp.Message)
				} else if code := status.Code(err); code == codes.DeadlineExceeded || code == codes.Canceled {
					return nil
				} else {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&req.Follow, "follow", "f", false, "Specify if the logs should be streamed.")
	return cmd
}
