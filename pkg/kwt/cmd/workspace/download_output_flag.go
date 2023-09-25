package workspace

import (
	"fmt"
	"strings"

	ctlwork "github.com/carvel-dev/kwt/pkg/kwt/workspace"
	"github.com/spf13/pflag"
)

type DownloadOutputFlag struct {
	outputs *[]ctlwork.DownloadOutput
}

var _ pflag.Value = &DownloadOutputFlag{}

func NewDownloadOutputFlag(outputs *[]ctlwork.DownloadOutput) *DownloadOutputFlag {
	return &DownloadOutputFlag{outputs}
}

func (s *DownloadOutputFlag) Set(val string) error {
	pieces := strings.SplitN(val, "=", 2)
	if len(pieces) != 2 {
		return fmt.Errorf("Expected input to be formatted as 'name=local-directory-path'")
	}

	if len(pieces[0]) == 0 {
		return fmt.Errorf("Expected input name to be non-empty")
	}

	name := pieces[0]

	if len(pieces[1]) == 0 {
		return fmt.Errorf("Expected input local directory path to be non-empty")
	}

	localPath := pieces[1]
	remotePath := ""

	if strings.Contains(localPath, ":") {
		pathPieces := strings.SplitN(localPath, ":", 2)
		if len(pathPieces) != 2 {
			return fmt.Errorf("Expected input path to be 'local-directory-path:relative-remote-path' or 'local-directory-path:/absolute-remote-path'")
		}

		localPath = pathPieces[0]
		remotePath = pathPieces[1]
	}

	*s.outputs = append(*s.outputs, ctlwork.DownloadOutput{ctlwork.Asset{Name: name, LocalDirPath: localPath, RemoteDirPath: remotePath}})

	return nil
}

func (s *DownloadOutputFlag) Type() string   { return "string" }
func (s *DownloadOutputFlag) String() string { return "" } // default for usage
