package mdns

import (
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type LocalIfaceMsgFilter struct {
	addrs     map[string]struct{}
	addrsLock sync.RWMutex

	logger Logger
	logTag string
}

var _ MsgFilter = &LocalIfaceMsgFilter{}

func NewLocalIfaceMsgFilter(logger Logger) *LocalIfaceMsgFilter {
	return &LocalIfaceMsgFilter{logger: logger, logTag: "mdns.LocalIfaceMsgFilter"}
}

func (f *LocalIfaceMsgFilter) UpdateIfacesContiniously() {
	for {
		_, err := f.updateIfacesOnce()
		if err != nil {
			f.logger.Debug(f.logTag, "Failed updating ifaces: %s", err)
		}
		time.Sleep(10 * time.Second)
	}
}

func (f *LocalIfaceMsgFilter) updateIfacesOnce() (map[string]struct{}, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	newAddrs := map[string]struct{}{}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil || ip == nil {
			continue
		}
		newAddrs[ip.String()] = struct{}{}
	}

	f.addrsLock.Lock()
	defer f.addrsLock.Unlock()

	if !reflect.DeepEqual(f.addrs, newAddrs) {
		f.addrs = newAddrs
		f.logger.Debug(f.logTag, "Iface addrs changed: %v", newAddrs)
	}

	return f.addrs, nil
}

func (f *LocalIfaceMsgFilter) AcceptMsg(msg *dns.Msg, srcAddr net.Addr) bool {
	f.logger.Debug(f.logTag, "Checking on: %s", srcAddr)

	ipStr, _, err := net.SplitHostPort(srcAddr.String())
	if err != nil {
		return false
	}

	f.addrsLock.RLock()
	numOfAddrs := len(f.addrs)
	_, found := f.addrs[ipStr]
	f.addrsLock.RUnlock()

	if numOfAddrs > 0 {
		return found
	}

	addrs, err := f.updateIfacesOnce()
	if err != nil {
		f.logger.Debug(f.logTag, "Failed updating ifaces: %s", err)
		return false
	}

	_, found = addrs[ipStr]
	return found
}
