package mdns

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

type MsgFilter interface {
	AcceptMsg(*dns.Msg, net.Addr) bool
}

type Server struct {
	handler dns.Handler
	filter  MsgFilter
	udpConn *net.UDPConn

	logTag string
	logger Logger
}

func NewServer(handler dns.Handler, filter MsgFilter, logger Logger) *Server {
	return &Server{handler: handler, filter: filter, logTag: "mdns.Server", logger: logger}
}

func (s *Server) Serve(startedCh chan struct{}) error {
	err := s.listen()
	if err != nil {
		return err
	}

	startedCh <- struct{}{}

	s.logger.Info(s.logTag, "Started mDNS resolver on %s", s.udpConn.LocalAddr())

	for {
		sess, err := newMsgSession(s.udpConn, s.filter)
		if err != nil {
			s.logger.Error(s.logTag, "Failed reading msg: %s", err)
			continue
		}
		if sess != nil {
			go s.handler.ServeDNS(&RespWriter{sess}, sess.msg)
		}
	}

	return nil
}

func (s *Server) Shutdown() error {
	return nil
}

func (s *Server) listen() error {
	udpConn, err := net.ListenMulticastUDP("udp4", nil, mdnsIPv4Addr)
	if err != nil {
		return fmt.Errorf("Listening multicast udp4: %s", err)
	}

	s.udpConn = udpConn

	return nil
}

type msgSession struct {
	msg     *dns.Msg
	srcAddr net.Addr
	udpConn *net.UDPConn
}

func newMsgSession(udpConn *net.UDPConn, filter MsgFilter) (*msgSession, error) {
	sess := &msgSession{}
	buf := make([]byte, 65536)

	n, srcAddr, err := udpConn.ReadFromUDP(buf)
	if err != nil {
		return nil, fmt.Errorf("Reading msg from conn: %s", err)
	}

	var msg dns.Msg

	err = msg.Unpack(buf[:n])
	if err != nil {
		if err != dns.ErrTruncated {
			return nil, fmt.Errorf("Unpacking msg: %s", err)
		}
	}

	// Bunch of empty DNS packets may come in... ignore them
	if len(msg.Question) == 0 {
		return nil, nil
	}

	if !filter.AcceptMsg(&msg, srcAddr) {
		return nil, nil
	}

	sess.msg = &msg
	sess.srcAddr = srcAddr
	sess.udpConn = udpConn

	return sess, nil
}

func (s *msgSession) WriteMsgMulticast(msg *dns.Msg) (int, error) {
	return s.writeMsgAddr(msg, mdnsIPv4Addr)
}

func (s *msgSession) WriteMsgUnicast(msg *dns.Msg) (int, error) {
	return s.writeMsgAddr(msg, s.srcAddr)
}

func (s *msgSession) writeMsgAddr(msg *dns.Msg, addr net.Addr) (int, error) {
	buf, err := msg.Pack()
	if err != nil {
		return 0, err
	}

	return s.udpConn.WriteTo(buf, addr)
}
