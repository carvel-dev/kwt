package dns

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

type (
	DomainsMapFunc     func() (map[string]IPResolver, error)
	DomainsChangedFunc func()
)

type DomainsMux struct {
	mux *dns.ServeMux

	mapFunc     DomainsMapFunc
	changedFunc DomainsChangedFunc
	prevDomains map[string]struct{}

	logTag string
	logger Logger
}

var _ dns.Handler = &DomainsMux{}

func NewDomainsMux(mux *dns.ServeMux, mapFunc DomainsMapFunc, changedFunc DomainsChangedFunc, logger Logger) *DomainsMux {
	return &DomainsMux{mux, mapFunc, changedFunc, map[string]struct{}{}, "dns.DomainsMux", logger}
}

func (m *DomainsMux) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m.mux.ServeDNS(w, r)
}

// UpdateContiniously and UpdateOnce are not thread safe
func (m *DomainsMux) UpdateContiniously() {
	for {
		err := m.UpdateOnce()
		if err != nil {
			m.logger.Debug(m.logTag, "Failed updating DNS domain handlers: %s", err)
		}
		time.Sleep(30 * time.Second)
	}
}

func (m *DomainsMux) UpdateOnce() error {
	domains, err := m.mapFunc()
	if err != nil {
		return fmt.Errorf("Fetching domains: %s", err)
	}

	m.logger.Debug(m.logTag, "Updating DNS domain handlers: %v", domains)

	changed := false

	for domain, resolver := range domains {
		if _, found := m.prevDomains[domain]; found {
			delete(m.prevDomains, domain)
		} else {
			m.logger.Info(m.logTag, "Registering %s->%s", domain, resolver)
			changed = true
		}
		m.mux.Handle(domain, NewCustomHandler(resolver, m.logger))
	}

	// Delete previously registered handlers that were not replaced
	for domain, _ := range m.prevDomains {
		m.logger.Info(m.logTag, "Unregistering %s", domain)
		changed = true
		m.mux.HandleRemove(domain)
	}

	m.prevDomains = map[string]struct{}{}

	// Record what was registered again
	for domain, _ := range domains {
		m.prevDomains[domain] = struct{}{}
	}

	if changed {
		m.changedFunc()
	}

	return nil
}
