package dns

import (
	"net"
	"time"

	"github.com/cppforlife/kwt/pkg/kwt/dnsutil"
	"github.com/miekg/dns"
)

type ForwardHandler struct {
	recursors RecursorPool

	nonScopedLogger Logger
	logTag          string
}

func NewForwardHandler(recursors RecursorPool, logger Logger) ForwardHandler {
	return ForwardHandler{
		recursors:       recursors,
		nonScopedLogger: logger,
		logTag:          "dns.ForwardHandler",
	}
}

func (r ForwardHandler) ServeDNS(responseWriter dns.ResponseWriter, request *dns.Msg) {
	logger := dnsutil.NewMsgPrefixedLogger(request, r.nonScopedLogger)

	if len(request.Question) == 0 {
		r.writeEmptyMessage(responseWriter, request, logger)
		return
	}

	logger.Debug(r.logTag, "Received query")

	t1 := time.Now()

	network := r.network(responseWriter)
	client := &dns.Client{Net: network, Timeout: 5 * time.Second, UDPSize: 65535}
	usedRecursor := ""

	err := r.recursors.PerformStrategically(func(recursor string) error {
		exchangeAnswer, _, exchangeErr := client.Exchange(request, recursor)
		if exchangeErr != nil && exchangeErr != dns.ErrTruncated {
			logger.Debug(r.logTag, "Failed recursing to %q: %s", recursor, exchangeErr)
			return exchangeErr
		}

		response := r.compressIfNeeded(responseWriter, request, exchangeAnswer, logger)

		writeErr := responseWriter.WriteMsg(response)
		if writeErr != nil {
			logger.Error(r.logTag, "Failed writing response: %s", writeErr)
		} else {
			usedRecursor = recursor
		}

		return nil
	})

	if err != nil {
		r.writeFailureMessage(responseWriter, request, logger)
		logger.Error(r.logTag, "Failed to recurse: %s", err)
	} else {
		logger.Info(r.logTag, "Answering via=%s (%s)", usedRecursor, time.Now().Sub(t1))
	}
}

func (r ForwardHandler) compressIfNeeded(
	responseWriter dns.ResponseWriter, request, response *dns.Msg, logger Logger) *dns.Msg {

	if _, ok := responseWriter.RemoteAddr().(*net.UDPAddr); ok {
		maxUDPSize := 512
		if opt := request.IsEdns0(); opt != nil {
			maxUDPSize = int(opt.UDPSize())
		}

		if response.Len() > maxUDPSize {
			logger.Debug(r.logTag, "Setting compress flag on msg id '%s'", request.Id)
			responseCopy := dns.Msg(*response)
			responseCopy.Compress = true
			return &responseCopy
		}
	}

	return response
}

func (ForwardHandler) network(responseWriter dns.ResponseWriter) string {
	network := "udp"
	if _, ok := responseWriter.RemoteAddr().(*net.TCPAddr); ok {
		network = "tcp"
	}
	return network
}

func (r ForwardHandler) writeFailureMessage(
	responseWriter dns.ResponseWriter, req *dns.Msg, logger Logger) {

	msg := &dns.Msg{}
	msg.SetReply(req)
	msg.SetRcode(req, dns.RcodeServerFailure)

	err := responseWriter.WriteMsg(msg)
	if err != nil {
		logger.Error(r.logTag, "Failed writing response: %s", err)
	}
}

func (r ForwardHandler) writeEmptyMessage(
	responseWriter dns.ResponseWriter, req *dns.Msg, logger Logger) {

	logger.Info(r.logTag, "received a request with no questions")

	msg := &dns.Msg{}
	msg.Authoritative = true
	msg.SetRcode(req, dns.RcodeSuccess)

	err := responseWriter.WriteMsg(msg)
	if err != nil {
		logger.Error(r.logTag, "Failed writing response: %s", err)
	}
}
