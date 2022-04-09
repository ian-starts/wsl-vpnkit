package virtualnetwork

import (
	"net"

	"github.com/containers/gvisor-tap-vsock/pkg/tap"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/pkg/errors"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func setupGateway(configuration *types.Configuration, networkSwitch *tap.Switch, stack *stack.Stack) error {
	tapEndpoint, err := tap.NewLinkEndpoint(configuration.Debug, configuration.MTU, configuration.GatewayMacAddress, configuration.GatewayIP, configuration.GatewayVirtualIPs)
	if err != nil {
		return errors.Wrap(err, "cannot create tap endpoint")
	}
	tapEndpoint.Connect(networkSwitch)
	networkSwitch.Connect(tapEndpoint)
	return createNIC(stack, 1, tapEndpoint, net.ParseIP(configuration.GatewayIP))
}
