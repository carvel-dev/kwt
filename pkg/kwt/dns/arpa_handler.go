package dns

import (
	"github.com/miekg/dns"
)

type DNSHandler interface {
	dns.Handler
}

type ArpaHandler struct {
	logger  Logger
	handler DNSHandler
	logTag  string
}

func NewArpaHandler(h DNSHandler, logger Logger) ArpaHandler {
	return ArpaHandler{
		handler: h,
		logger:  logger,
		logTag:  "dns.ArpaHandler",
	}
}

func (a ArpaHandler) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := &dns.Msg{}

	// TODO correct?
	msg.Authoritative = true
	msg.RecursionAvailable = false

	if len(req.Question) == 0 {
		msg.SetRcode(req, dns.RcodeSuccess)

		err := w.WriteMsg(msg)
		if err != nil {
			a.logger.Error(a.logTag, "Failed writing response: %s", err)
		}

		return
	}

	a.handler.ServeDNS(w, req)
}
