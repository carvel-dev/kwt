package forwarder

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cppforlife/kwt/pkg/kwt/net/pf"
)

// Piece of example output: 'Token : 13073750315387878699'
var enableTokenRegexp = regexp.MustCompile(`Token : (.+)`)

type Pfctl struct {
	name string
	opts PfctlOpts

	enableToken string

	logTag string
	logger Logger
}

var _ Forwarder = &Pfctl{}

type PfctlOpts struct {
	DstTCPPort    int
	DstDNSUDPPort int

	// 0 is root, so root produced traffic wont go through proxy by default
	ProcessGroupID int
}

func NewPfctl(opts PfctlOpts, logger Logger) *Pfctl {
	return &Pfctl{
		name: fmt.Sprintf("kwt-tcp-%d", opts.DstTCPPort),
		opts: opts,

		logTag: "Pfctl",
		logger: logger,
	}
}

func (f *Pfctl) Add(subnets []net.IPNet, dnsIPs []net.IP) error {
	// TODO check if disabled?
	// TODO local loopback

	err := f.addAnchorIfNotExists()
	if err != nil {
		return err
	}

	err = f.addRules(subnets, dnsIPs)
	if err != nil {
		return err
	}

	output, err := f.run([]string{"-E"}, nil)
	if err != nil {
		return err
	}

	matches := enableTokenRegexp.FindStringSubmatch(output)
	if len(matches) == 2 {
		f.enableToken = matches[1]
	}

	return nil
}

func (f *Pfctl) Reset() error {
	_, err := f.run([]string{"-a", f.name, "-F", "all"}, nil)
	if err != nil {
		return err
	}

	err = f.deleteAnchorIfNotExists()
	if err != nil {
		return err
	}

	if len(f.enableToken) > 0 {
		_, err = f.run([]string{"-X", f.enableToken}, nil)
		if err != nil {
			return err
		}

		f.enableToken = ""
	}

	return nil
}

func (f *Pfctl) addAnchorIfNotExists() error {
	output, err := f.run([]string{"-s", "all"}, nil)
	if err != nil {
		return err
	}

	pfctl, err := pf.NewPfctl()
	if err != nil {
		return err
	}

	defer pfctl.Close()

	anchorTypes := map[pf.RuleAction]string{
		pf.PF_RDR:  "rdr-anchor",
		pf.PF_PASS: "anchor",
	}

	for ruleAction, anchorType := range anchorTypes {
		// newline in the beginning to avoid partial match
		if !strings.Contains(output, fmt.Sprintf("\n%s \"%s\"", anchorType, f.name)) {
			err := pfctl.AddAnchorIfNotExist(f.name, ruleAction)
			if err != nil {
				return fmt.Errorf("Adding %s: %s", anchorType, err)
			}
		}
	}

	return nil
}

func (f *Pfctl) deleteAnchorIfNotExists() error {
	pfctl, err := pf.NewPfctl()
	if err != nil {
		return err
	}

	defer pfctl.Close()

	return pfctl.DeleteAnchorIfExists(f.name, []pf.RuleAction{pf.PF_RDR, pf.PF_PASS})
}

func (f *Pfctl) addRules(subnets []net.IPNet, dnsIPs []net.IP) error {
	rules := fmt.Sprintf(`
table <forward_subnets> {|subnets|}
table <dns_servers> {|dnsservers|}
rdr pass on lo0 inet proto tcp to <forward_subnets> -> 127.0.0.1 port |tcpport|
rdr pass on lo0 inet proto udp to <dns_servers> port 53 -> 127.0.0.1 port |dnsudpport|
pass out route-to lo0 inet proto tcp to <forward_subnets> keep state |excludegroup|
pass out route-to lo0 inet proto udp to <dns_servers> port 53 keep state |excludegroup|
`)

	subnetsStrs := []string{"!127.0.0.1/32"}
	for _, subnet := range subnets {
		subnetsStrs = append(subnetsStrs, subnet.String())
	}

	var dnsServerStrs []string
	for _, ip := range dnsIPs {
		dnsServerStrs = append(dnsServerStrs, ip.String())
	}

	rules = strings.Replace(rules, "|tcpport|", strconv.Itoa(f.opts.DstTCPPort), -1)
	rules = strings.Replace(rules, "|dnsudpport|", strconv.Itoa(f.opts.DstDNSUDPPort), -1)
	rules = strings.Replace(rules, "|subnets|", strings.Join(subnetsStrs, ","), -1)
	rules = strings.Replace(rules, "|dnsservers|", strings.Join(dnsServerStrs, ","), -1)
	rules = strings.Replace(rules, "|excludegroup|", fmt.Sprintf("group {!=%d}", f.opts.ProcessGroupID), -1)

	f.logger.Debug(f.logTag, "Will run pfctl with following rules: %s", rules)

	_, err := f.run([]string{"-a", f.name, "-f", "-"}, strings.NewReader(rules))

	return err
}

func (f *Pfctl) run(args []string, stdin io.Reader) (string, error) {
	cmd := exec.Command("pfctl", args...)
	cmd.Stdin = stdin

	argsDesc := strings.Join(args, " ")

	f.logger.Debug(f.logTag, "Running 'pfctl %s'", argsDesc)

	output, err := cmd.CombinedOutput()
	if err != nil {
		f.logger.Debug(f.logTag, "Failed, error: %s, output: %s", err, output)
		return "", fmt.Errorf("Running 'pfctl %s': %s", argsDesc, err)
	} else {
		f.logger.Debug(f.logTag, "Succeeded, output: %s", output)
	}

	return string(output), nil
}
