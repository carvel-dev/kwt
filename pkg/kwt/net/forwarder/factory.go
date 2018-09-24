package forwarder

import (
	"fmt"
	"runtime"
)

const (
	osDarwin = "darwin"
	osLinux  = "linux"
)

type Factory struct {
	processGroupID int
	logger         Logger
}

func NewFactory(processGroupID int, logger Logger) Factory {
	return Factory{processGroupID, logger}
}

type ForwarderOpts struct {
	DstTCPPort    int
	DstDNSUDPPort int
}

func (f Factory) NewForwarder(opts ForwarderOpts) (Forwarder, error) {
	os := runtime.GOOS

	switch os {
	case osLinux:
		opts := IptablesOpts{
			DstTCPPort:     opts.DstTCPPort,
			DstDNSUDPPort:  opts.DstDNSUDPPort,
			ProcessGroupID: f.processGroupID,
		}
		return NewIptables(opts, NewOsCmdExecutor(f.logger), f.logger), nil

	case osDarwin:
		opts := PfctlOpts{
			DstTCPPort:     opts.DstTCPPort,
			DstDNSUDPPort:  opts.DstDNSUDPPort,
			ProcessGroupID: f.processGroupID,
		}
		return NewPfctl(opts, f.logger), nil

	default:
		return nil, fmt.Errorf("OS '%s' is not supported for connection forwarding", os)
	}
}

func (f Factory) NewOriginalDstResolver() (OriginalDstResolver, error) {
	os := runtime.GOOS

	switch os {
	case osLinux:
		return LinuxOriginalDstResolver{}, nil

	case osDarwin:
		return NewPfctlResolver(f.logger)

	default:
		return nil, fmt.Errorf("OS '%s' is not supported for original destination resolution", os)
	}
}
