package virtualnetwork

import (
	"net"
	"sync"

	"github.com/containers/gvisor-tap-vsock/pkg/services/forwarder"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

func addNAT(configuration *types.Configuration, s *stack.Stack) {
	var natLock sync.Mutex
	translation := parseNATTable(configuration)
	tcpForwarder := forwarder.TCP(s, translation, &natLock)
	s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)
	udpForwarder := forwarder.UDP(s, translation, &natLock)
	s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)
}

func parseNATTable(configuration *types.Configuration) map[tcpip.Address]tcpip.Address {
	translation := make(map[tcpip.Address]tcpip.Address)
	for source, destination := range configuration.NAT {
		translation[tcpip.Address(net.ParseIP(source).To4())] = tcpip.Address(net.ParseIP(destination).To4())
	}
	return translation
}
