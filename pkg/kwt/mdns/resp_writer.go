package mdns

import (
	"net"

	"github.com/miekg/dns"
)

type RespWriter struct {
	sess *msgSession
}

var _ dns.ResponseWriter = &RespWriter{}

func (w RespWriter) LocalAddr() net.Addr {
	panic("Not implemented")
	return nil
}

func (w RespWriter) RemoteAddr() net.Addr {
	panic("Not implemented")
	return nil
}

func (w RespWriter) WriteMsg(msg *dns.Msg) error {
	// https://tools.ietf.org/html/rfc6762#section-5.4
	replyUnicast := msg.Question[0].Qclass&32768 > 0

	// Clear out question so that no one answers this again
	msg.Question = nil

	if replyUnicast {
		_, err := w.sess.WriteMsgUnicast(msg)
		return err
	}

	_, err := w.sess.WriteMsgMulticast(msg)
	return err
}

func (w RespWriter) Write([]byte) (int, error) {
	panic("Not implemented")
	return 0, nil
}

func (w RespWriter) Close() error {
	panic("Not implemented")
	return nil
}

func (w RespWriter) TsigStatus() error {
	panic("Not implemented")
	return nil
}

func (w RespWriter) TsigTimersOnly(bool) {
	panic("Not implemented")
}

func (w RespWriter) Hijack() {
	panic("Not implemented")
}
