package dns

import (
	"github.com/miekg/dns"
)

type Factory struct{}

type BuildOpts struct {
	ListenAddrs   []string // include port
	RecursorAddrs []string // include port

	DomainsMapFunc     DomainsMapFunc
	DomainsChangedFunc DomainsChangedFunc
}

func NewFactory() Factory { return Factory{} }

func (f Factory) Build(opts BuildOpts, logger Logger) (Server, error) {
	recursorPool := NewFailoverRecursorPool(opts.RecursorAddrs, logger)
	forwardHandler := NewForwardHandler(recursorPool, logger)
	arpaHandler := NewArpaHandler(forwardHandler, logger)

	mux := dns.NewServeMux()
	mux.Handle("arpa.", arpaHandler)
	mux.Handle(".", forwardHandler)

	domainsMux := NewDomainsMux(mux, opts.DomainsMapFunc, opts.DomainsChangedFunc, logger)

	err := domainsMux.UpdateOnce()
	if err != nil {
		return Server{}, err
	}

	go domainsMux.UpdateContiniously()

	servers := []*dns.Server{}

	for _, addr := range opts.ListenAddrs {
		servers = append(servers,
			&dns.Server{Addr: addr, Net: "tcp", Handler: domainsMux},
			&dns.Server{Addr: addr, Net: "udp", Handler: domainsMux, UDPSize: 65535},
		)
	}

	return NewServer(servers, logger), nil
}
