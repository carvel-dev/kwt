package forwarder

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cppforlife/kwt/pkg/kwt/net/pf"
)

type PfctlResolver struct {
	pfctl *pf.Pfctl

	logTag string
	logger Logger
}

var _ OriginalDstResolver = PfctlResolver{}

func NewPfctlResolver(logger Logger) (PfctlResolver, error) {
	pfctl, err := pf.NewPfctl()
	if err != nil {
		return PfctlResolver{}, err
	}

	return PfctlResolver{pfctl, "PfctlResolver", logger}, nil
}

func (r PfctlResolver) GetOrigIPPort(conn net.Conn) (net.IP, int, error) {
	srcIPStr, srcPortStr, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return nil, 0, fmt.Errorf("Parsing src addr: %s", err)
	}

	dstIPStr, dstPortStr, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return nil, 0, fmt.Errorf("Parsing dst addr: %s", err)
	}

	srcPort, err := strconv.Atoi(srcPortStr)
	if err != nil {
		return nil, 0, fmt.Errorf("Parsing src port: %s", err)
	}

	dstPort, err := strconv.Atoi(dstPortStr)
	if err != nil {
		return nil, 0, fmt.Errorf("Parsing dst port: %s", err)
	}

	opts := PfctlResolverOpts{
		SrcIP:   net.ParseIP(srcIPStr),
		SrcPort: int32(srcPort),
		DstIP:   net.ParseIP(dstIPStr),
		DstPort: int32(dstPort),
	}

	return r.GetOrigIPPortWithOpts(opts)
}

type PfctlResolverOpts struct {
	SrcIP   net.IP
	SrcPort int32
	DstIP   net.IP
	DstPort int32
}

func (r PfctlResolver) GetOrigIPPortWithOpts(opts PfctlResolverOpts) (net.IP, int, error) {
	return r.pfctl.LookUpNAT(pf.LookUpNATOpts{
		SrcIP:   opts.SrcIP,
		SrcPort: opts.SrcPort,
		DstIP:   opts.DstIP,
		DstPort: opts.DstPort,
	})
}
