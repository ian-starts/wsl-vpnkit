package virtualnetwork

import (
	"net"

	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/arp"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

func createStack(parsedSubnet *net.IPNet) (*stack.Stack, error) {
	s := stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			arp.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
		},
	})

	subnet, err := tcpip.NewSubnet(tcpip.Address(parsedSubnet.IP), tcpip.AddressMask(parsedSubnet.Mask))
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse subnet")
	}
	s.SetRouteTable([]tcpip.Route{
		{
			Destination: subnet,
			Gateway:     "",
			NIC:         1,
		},
	})

	return s, nil
}

func createNIC(s *stack.Stack, id tcpip.NICID, endpoint stack.LinkEndpoint, ip net.IP) error {
	if err := s.CreateNIC(id, endpoint); err != nil {
		return errors.New(err.String())
	}

	if err := s.AddProtocolAddress(id, tcpip.ProtocolAddress{
		Protocol:          ipv4.ProtocolNumber,
		AddressWithPrefix: tcpip.Address(ip.To4()).WithPrefix(),
	}, stack.AddressProperties{}); err != nil {
		return errors.New(err.String())
	}

	s.SetSpoofing(id, true)
	s.SetPromiscuousMode(id, true)
	return nil
}
