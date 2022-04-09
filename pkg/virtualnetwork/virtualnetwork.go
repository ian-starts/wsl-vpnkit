package virtualnetwork

import (
	"context"
	"net"

	"github.com/containers/gvisor-tap-vsock/pkg/tap"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type VirtualNetwork interface {
	Accept(ctx context.Context, conn net.Conn) error
	DialUDP(ip net.IP, port uint16) (net.PacketConn, error)
	ListenTCP(ip net.IP, port uint16) (net.Listener, error)
}

type virtualNetwork struct {
	configuration *types.Configuration
	stack         *stack.Stack
	networkSwitch *tap.Switch
}

func New(configuration *types.Configuration) (VirtualNetwork, error) {
	_, subnet, err := net.ParseCIDR(configuration.Subnet)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse subnet cidr")
	}

	stack, err := createStack(subnet)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create network stack")
	}

	networkSwitch := tap.NewSwitch(configuration.Debug, configuration.MTU, configuration.Protocol)

	err = setupGateway(configuration, networkSwitch, stack)
	if err != nil {
		return nil, err
	}

	addNAT(configuration, stack)

	err = dhcpServer(configuration, stack)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add dhcp service")
	}

	return &virtualNetwork{
		configuration: configuration,
		stack:         stack,
		networkSwitch: networkSwitch,
	}, nil
}

func (n *virtualNetwork) Accept(ctx context.Context, conn net.Conn) error {
	return n.networkSwitch.Accept(ctx, conn)
}

func (n *virtualNetwork) DialUDP(ip net.IP, port uint16) (net.PacketConn, error) {
	return gonet.DialUDP(n.stack, &tcpip.FullAddress{
		NIC:  1,
		Addr: tcpip.Address(ip.To4()),
		Port: uint16(53),
	}, nil, ipv4.ProtocolNumber)
}

func (n *virtualNetwork) ListenTCP(ip net.IP, port uint16) (net.Listener, error) {
	return gonet.ListenTCP(n.stack, tcpip.FullAddress{
		NIC:  1,
		Addr: tcpip.Address(ip.To4()),
		Port: uint16(port),
	}, ipv4.ProtocolNumber)
}
