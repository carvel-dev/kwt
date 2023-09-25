package dns

import (
	"net"
	"time"

	"github.com/carvel-dev/kwt/pkg/kwt/dnsutil"
	"github.com/miekg/dns"
)

type IPResolver interface {
	ResolveIPv4(string) ([]net.IP, bool, error)
	ResolveIPv6(string) ([]net.IP, bool, error)
}

type CustomHandler struct {
	ipResolver IPResolver

	nonScopedLogger Logger
	logTag          string
}

func NewCustomHandler(ipResolver IPResolver, logger Logger) CustomHandler {
	return CustomHandler{
		ipResolver: ipResolver,

		nonScopedLogger: logger,
		logTag:          "dns.CustomHandler",
	}
}

func (d CustomHandler) ServeDNS(responseWriter dns.ResponseWriter, requestMsg *dns.Msg) {
	logger := dnsutil.NewMsgPrefixedLogger(requestMsg, d.nonScopedLogger)

	logger.Debug(d.logTag, "Received query")

	msg := &dns.Msg{}
	t1 := time.Now()

	if len(requestMsg.Question) > 0 {
		question := requestMsg.Question[0]

		switch question.Qtype {
		case dns.TypeA, dns.TypeANY:
			ips, resolved, err := d.ipResolver.ResolveIPv4(question.Name)
			if !resolved || err != nil {
				msg.SetRcode(requestMsg, dns.RcodeServerFailure)
			} else {
				msg.SetRcode(requestMsg, dns.RcodeSuccess)

				for _, ip := range ips {
					msg.Answer = append(msg.Answer, &dns.A{
						Hdr: dns.RR_Header{
							Name:   question.Name,
							Rrtype: dns.TypeA,
							Class:  dns.ClassINET,
							Ttl:    0, // OS X seems to have min TTL of 17s
						},
						A: ip,
					})
				}
			}

		case dns.TypeAAAA:
			// TODO IPv6
			msg.SetRcode(requestMsg, dns.RcodeSuccess)

		case dns.TypeMX:
			msg.SetRcode(requestMsg, dns.RcodeSuccess)

		default:
			msg.SetRcode(requestMsg, dns.RcodeServerFailure)
		}
	}

	msg.Authoritative = true
	msg.RecursionAvailable = true

	err := responseWriter.WriteMsg(msg)
	if err != nil {
		logger.Error(d.logTag, "Failed writing response: %s", err)
	} else {
		logger.Info(d.logTag, "Answering rcode=%d (%s)", msg.Rcode, time.Now().Sub(t1))
	}
}
