package dns

import (
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
	"strings"
)

var nameserverRegex = regexp.MustCompile("^\\s*nameserver\\s+(\\S+)$")

type ResolvConf struct{}

func NewResolvConf() ResolvConf {
	return ResolvConf{}
}

func (r ResolvConf) Nameservers() ([]net.IP, error) {
	bytes, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("Reading DNS nameservers: %s", err)
	}

	lines := strings.Split(string(bytes), "\n")

	var ips []net.IP

	for _, line := range lines {
		submatch := nameserverRegex.FindAllStringSubmatch(line, 1)
		if len(submatch) > 0 {
			ip := net.ParseIP(submatch[0][1])
			if ip != nil {
				ips = append(ips, ip)
			}
		}
	}

	return ips, nil
}
