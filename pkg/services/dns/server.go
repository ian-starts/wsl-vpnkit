package dns

import (
	"net"

	"github.com/miekg/dns"
	"github.com/sakai135/wsl-vpnkit/pkg/virtualnetwork"
)

func NewServer(vn virtualnetwork.VirtualNetwork, handler dns.Handler, activateAndServe func(l net.Listener, p net.PacketConn, handler dns.Handler) error) *server {
	return &server{vn: vn, handler: handler, activateAndServe: activateAndServe}
}

type server struct {
	activateAndServe func(l net.Listener, p net.PacketConn, handler dns.Handler) error
	handler          dns.Handler
	vn               virtualnetwork.VirtualNetwork
}

func (s *server) StartTCP(ip net.IP, port uint16) error {
	listener, err := s.vn.ListenTCP(ip, port)
	if err != nil {
		return err
	}
	return s.activateAndServe(listener, nil, s.handler)
}

func (s *server) StartUDP(ip net.IP, port uint16) error {
	conn, err := s.vn.DialUDP(ip, port)
	if err != nil {
		return err
	}
	return s.activateAndServe(nil, conn, s.handler)
}
