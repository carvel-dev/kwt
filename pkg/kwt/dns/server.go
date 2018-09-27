package dns

import (
	"fmt"
	"net"
	"sync"

	"github.com/miekg/dns"
)

type Server struct {
	servers    []*dns.Server
	shutdownCh chan struct{}

	logTag string
	logger Logger
}

func NewServer(servers []*dns.Server, logger Logger) Server {
	return Server{
		servers:    servers,
		shutdownCh: make(chan struct{}),

		logTag: "dns.Server",
		logger: logger,
	}
}

func (s Server) Serve(startedCh chan struct{}) error {
	errCh := make(chan error)

	// Cannot use (*dns.Server).ListenAndServer unfortunately
	err := s.listen()
	if err != nil {
		return err
	}

	startedCh <- struct{}{}

	s.logger.Info(s.logTag, "Started DNS server on %s (TCP) and %s (UDP)", s.TCPAddr(), s.UDPAddr())

	for _, srv := range s.servers {
		go func(srv *dns.Server) {
			errCh <- srv.ActivateAndServe()
		}(srv)
	}

	select {
	case err := <-errCh:
		return fmt.Errorf("Serving: %s", err)
	case <-s.shutdownCh:
		return s.shutdown()
	}
}

func (s Server) TCPAddr() net.Addr {
	for _, srv := range s.servers {
		switch {
		case srv.Listener != nil:
			return srv.Listener.Addr()
		}
	}

	panic("Unknown TCP DNS server address")
}

func (s Server) UDPAddr() net.Addr {
	for _, srv := range s.servers {
		switch {
		case srv.PacketConn != nil:
			if conn, ok := srv.PacketConn.(*net.UDPConn); ok {
				return conn.LocalAddr()
			}
		}
	}

	panic("Unknown UDP DNS server address")
}

func (s Server) Shutdown() error {
	close(s.shutdownCh)
	return nil
}

func (s Server) shutdown() error {
	errCh := make(chan error, len(s.servers))

	wg := &sync.WaitGroup{}
	wg.Add(len(s.servers))

	for _, server := range s.servers {
		go func(server *dns.Server) {
			errCh <- server.Shutdown()
			wg.Done()
		}(server)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s Server) listen() error {
	// Adapted from https://github.com/miekg/dns/blob/e875a31a5cfbcd646131e625f24818cbae228913/server.go#L407
	for _, srv := range s.servers {
		if srv.Listener != nil || srv.PacketConn != nil {
			continue
		}

		addr := srv.Addr
		if addr == "" {
			addr = ":domain"
		}
		if srv.UDPSize == 0 {
			srv.UDPSize = dns.MinMsgSize
		}

		switch srv.Net {
		case "tcp", "tcp4", "tcp6":
			a, err := net.ResolveTCPAddr(srv.Net, addr)
			if err != nil {
				return err
			}
			l, err := net.ListenTCP(srv.Net, a)
			if err != nil {
				return err
			}
			srv.Listener = l

		case "udp", "udp4", "udp6":
			a, err := net.ResolveUDPAddr(srv.Net, addr)
			if err != nil {
				return err
			}
			l, err := net.ListenUDP(srv.Net, a)
			if err != nil {
				return err
			}
			srv.PacketConn = l

		default:
			return fmt.Errorf("Unknown net '%s'", srv.Net)
		}
	}

	return nil
}
