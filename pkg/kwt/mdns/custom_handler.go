package mdns

import (
	"net"
	"time"

	"github.com/carvel-dev/kwt/pkg/kwt/dnsutil"
	"github.com/miekg/dns"
)

const (
	cacheFlushBit uint16 = 1 << 15
	ttl           uint32 = 5
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
		logTag:          "mdns.CustomHandler",
	}
}

type resolutionResult struct {
	IPs      []net.IP
	Resolved bool
	Err      error
}

func (d CustomHandler) ServeDNS(responseWriter dns.ResponseWriter, msg *dns.Msg) {
	logger := dnsutil.NewMsgPrefixedLogger(msg, d.nonScopedLogger)

	logger.Debug(d.logTag, "Received query")

	t1 := time.Now()

	// Clear out potentially filled info
	msg.Answer = []dns.RR{}
	msg.Ns = []dns.RR{}
	msg.Extra = []dns.RR{}

	authoritative := false

	for _, question := range msg.Question {
		var v4, v6 resolutionResult

		switch question.Qtype {
		case dns.TypeA, dns.TypeAAAA: // TODO dns.TypeANY?
			v4.IPs, v4.Resolved, v4.Err = d.ipResolver.ResolveIPv4(question.Name)
			if v4.Err != nil {
				logger.Error(d.logTag, "Failed resolving IPv4 question %s: %s", question.Name, v4.Err)
				return
			}

			v6.IPs, v6.Resolved, v6.Err = d.ipResolver.ResolveIPv6(question.Name)
			if v6.Err != nil {
				logger.Error(d.logTag, "Failed resolving IPv6 question %s: %s", question.Name, v6.Err)
				return
			}

			authoritative = authoritative || v4.Resolved || v6.Resolved

			// Opportunistically add both types of IPs
			for _, ip := range v4.IPs {
				msg.Answer = append(msg.Answer, d.recA(question, ip))
			}

			for _, ip := range v6.IPs {
				msg.Answer = append(msg.Answer, d.recAAAA(question, ip))
			}

		default:
			// do nothing
		}

		if authoritative {
			// Negative response for types that do no have IPs (eg commonly IPv6)
			msg.Extra = append(msg.Extra, d.recNSEC(question, v4, v6))
		}
	}

	if !authoritative {
		logger.Debug(d.logTag, "Skipping answering since not authoritative")
		return
	}

	msg.MsgHdr.Response = true
	msg.MsgHdr.Authoritative = true

	err := responseWriter.WriteMsg(msg)
	if err != nil {
		logger.Error(d.logTag, "Failed writing resp: %s", err.Error())
	} else {
		logger.Info(d.logTag, "Answering (%s)", time.Now().Sub(t1))
	}
}

func (d CustomHandler) recA(question dns.Question, ip net.IP) *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{
			Name:   question.Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET | cacheFlushBit,
			Ttl:    ttl,
		},
		A: ip,
	}
}

func (d CustomHandler) recAAAA(question dns.Question, ip net.IP) *dns.AAAA {
	return &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   question.Name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET | cacheFlushBit,
			Ttl:    ttl,
		},
		AAAA: ip,
	}
}

// Includes all DNS types that DO have records to return.
// https://tools.ietf.org/html/rfc6762#section-6.1
func (d CustomHandler) recNSEC(question dns.Question, v4, v6 resolutionResult) *dns.NSEC {
	nsecTypeBitMap := []uint16{}

	if len(v4.IPs) > 0 {
		nsecTypeBitMap = append(nsecTypeBitMap, dns.TypeA)
	}

	if len(v6.IPs) > 0 {
		nsecTypeBitMap = append(nsecTypeBitMap, dns.TypeAAAA)
	}

	return &dns.NSEC{
		Hdr: dns.RR_Header{
			Name:   question.Name,
			Rrtype: dns.TypeNSEC,
			Class:  dns.ClassINET | cacheFlushBit,
			Ttl:    ttl,
		},
		NextDomain: question.Name,
		TypeBitMap: nsecTypeBitMap,
	}
}
