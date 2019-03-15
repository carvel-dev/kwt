package workspace

import (
	"fmt"
	"strings"

	ctlwork "github.com/k14s/kwt/pkg/kwt/workspace"
	"github.com/spf13/pflag"
)

type UploadInputFlag struct {
	inputs *[]ctlwork.UploadInput
}

var _ pflag.Value = &UploadInputFlag{}

func NewUploadInputFlag(inputs *[]ctlwork.UploadInput) *UploadInputFlag {
	return &UploadInputFlag{inputs}
}

func (s *UploadInputFlag) Set(val string) error {
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

	*s.inputs = append(*s.inputs, ctlwork.UploadInput{ctlwork.Asset{Name: name, LocalDirPath: localPath, RemoteDirPath: remotePath}})

	return nil
}

func (s *UploadInputFlag) Type() string   { return "string" }
func (s *UploadInputFlag) String() string { return "" } // default for usage
