package workspace

import (
	"time"

	"k8s.io/client-go/rest"
)

type Workspace interface {
	Name() string
	Image() string
	State() string
	CreationTime() time.Time

	Ports() []string
	Privileged() bool

	LastUsedTime() time.Time
	MarkUse() error

	AltNames() []string
	AddAltName(string) error

	WaitForStart(chan struct{}) error

	Enter() error
	Execute(ExecuteOpts, *rest.Config) error

	Upload(UploadInput, *rest.Config) error // TODO remove rest.Config
	Download(DownloadOutput, *rest.Config) error

	Delete() error
}
