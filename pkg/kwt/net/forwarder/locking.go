package forwarder

import (
	"net"
	"sync"
)

type Locking struct {
	lock      sync.Mutex
	forwarder Forwarder
}

var _ Forwarder = &Locking{}

func NewLocking() *Locking {
	return &Locking{}
}

func (f *Locking) SetForwarder(forwarder Forwarder) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.forwarder = forwarder
}

func (f *Locking) CheckPrereqs() error {
	f.lock.Lock()
	defer f.lock.Unlock()

	return f.forwarder.CheckPrereqs()
}

func (f *Locking) Add(subnets []net.IPNet, dnsIPs []net.IP) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	return f.forwarder.Add(subnets, dnsIPs)
}

func (f *Locking) Reset() error {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.forwarder != nil {
		return f.forwarder.Reset()
	}
	return nil
}
