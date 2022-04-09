package dns

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"
)

func TestStartTCP_Success(t *testing.T) {
	vn := &mockVirtualNetwork{IP: net.ParseIP("192.168.1.1"), Port: 53}

	server := NewServer(vn, &mockHandler{}, mockActivateAndServe)
	err := server.StartTCP(net.ParseIP("192.168.1.1"), 53)
	if err != nil {
		t.Error("Expected err to be nil")
	}
}

func TestStartTCP_Error(t *testing.T) {
	vn := &mockVirtualNetwork{IP: net.ParseIP("192.168.1.1"), Port: 53}

	server := NewServer(vn, &mockHandler{}, mockActivateAndServe)
	err := server.StartTCP(net.ParseIP("192.168.1.2"), 53)
	if err == nil {
		t.Error("Expected err to not be nil")
	}
}

func TestStartUDP_Success(t *testing.T) {
	vn := &mockVirtualNetwork{IP: net.ParseIP("192.168.1.1"), Port: 53}

	server := NewServer(vn, &mockHandler{}, mockActivateAndServe)
	err := server.StartUDP(net.ParseIP("192.168.1.1"), 53)
	if err != nil {
		t.Error("Expected err to be nil")
	}
}

func TestStartUDP_Error(t *testing.T) {
	vn := &mockVirtualNetwork{IP: net.ParseIP("192.168.1.1"), Port: 53}

	server := NewServer(vn, &mockHandler{}, mockActivateAndServe)
	err := server.StartUDP(net.ParseIP("192.168.1.2"), 53)
	if err == nil {
		t.Error("Expected err to not be nil")
	}
}

type mockListener struct {
}

func (l *mockListener) Accept() (net.Conn, error) {
	return nil, nil
}

func (l *mockListener) Close() error {
	return nil
}

func (l *mockListener) Addr() net.Addr {
	return nil
}

type mockPacketConn struct {
}

func (c *mockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return 0, nil, nil
}

func (c *mockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return 0, nil
}

func (c *mockPacketConn) Close() error {
	return nil
}

func (c *mockPacketConn) LocalAddr() net.Addr {
	return nil
}

func (c *mockPacketConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *mockPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *mockPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type mockVirtualNetwork struct {
	IP   net.IP
	Port uint16
}

func (vn *mockVirtualNetwork) Accept(ctx context.Context, conn net.Conn) error {
	return nil
}

func (vn *mockVirtualNetwork) DialUDP(ip net.IP, port uint16) (net.PacketConn, error) {
	if !cmp.Equal(vn.IP, ip) {
		return nil, errors.New("unexpected ip")
	}
	if vn.Port != port {
		return nil, errors.New("unexpected port")
	}
	return &mockPacketConn{}, nil
}

func (vn *mockVirtualNetwork) ListenTCP(ip net.IP, port uint16) (net.Listener, error) {
	if !cmp.Equal(vn.IP, ip) {
		return nil, errors.New("unexpected ip")
	}
	if vn.Port != port {
		return nil, errors.New("unexpected port")
	}
	return &mockListener{}, nil
}

type mockHandler struct {
}

func (h *mockHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
}

func mockActivateAndServe(l net.Listener, p net.PacketConn, handler dns.Handler) error {
	return nil
}
