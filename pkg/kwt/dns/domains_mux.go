package dns

import (
	"time"

	"github.com/miekg/dns"
)

type DomainsFunc func() (map[string]IPResolver, error)

type DomainsMux struct {
	mux         *dns.ServeMux
	domainsFunc DomainsFunc
	prevDomains map[string]struct{}

	logTag string
	logger Logger
}

var _ dns.Handler = &DomainsMux{}

func NewDomainsMux(mux *dns.ServeMux, domainsFunc DomainsFunc, logger Logger) *DomainsMux {
	return &DomainsMux{mux, domainsFunc, map[string]struct{}{}, "dns.DomainsMux", logger}
}

func (m *DomainsMux) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m.mux.ServeDNS(w, r)
}

func (m *DomainsMux) UpdateDomainsContiniously() {
	for {
		err := m.updateDomainsOnce()
		if err != nil {
			m.logger.Debug(m.logTag, "Failed updating DNS domain handlers: %s", err)
		}
		time.Sleep(30 * time.Second)
	}
}

func (m *DomainsMux) updateDomainsOnce() error {
	domains, err := m.domainsFunc()
	if err != nil {
		return err
	}

	m.logger.Debug(m.logTag, "Updating DNS domain handlers: %v", domains)

	for domain, resolver := range domains {
		if _, found := m.prevDomains[domain]; found {
			delete(m.prevDomains, domain)
		} else {
			m.logger.Info(m.logTag, "Registering %s->%s", domain, resolver)
		}
		m.mux.Handle(domain, NewCustomHandler(resolver, m.logger))
	}

	// Delete previously registered handlers that were not replaced
	for domain, _ := range m.prevDomains {
		m.logger.Info(m.logTag, "Unregistering %s", domain)
		m.mux.HandleRemove(domain)
	}

	m.prevDomains = map[string]struct{}{}

	// Record what was registered again
	for domain, _ := range domains {
		m.prevDomains[domain] = struct{}{}
	}

	return nil
}
