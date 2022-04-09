package virtualnetwork

import (
	"net"

	"github.com/containers/gvisor-tap-vsock/pkg/services/dhcp"
	"github.com/containers/gvisor-tap-vsock/pkg/tap"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func dhcpServer(configuration *types.Configuration, s *stack.Stack) error {
	_, subnet, err := net.ParseCIDR(configuration.Subnet)
	if err != nil {
		return errors.Wrap(err, "cannot parse subnet cidr")
	}

	ipPool := tap.NewIPPool(subnet)
	ipPool.Reserve(net.ParseIP(configuration.GatewayIP), configuration.GatewayMacAddress)
	for ip, mac := range configuration.DHCPStaticLeases {
		ipPool.Reserve(net.ParseIP(ip), mac)
	}

	server, err := dhcp.New(configuration, s, ipPool)
	if err != nil {
		return err
	}
	go func() {
		log.Error(server.Serve())
	}()
	return nil
}
