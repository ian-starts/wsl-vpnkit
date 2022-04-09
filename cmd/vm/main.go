package main

import (
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sakai135/wsl-vpnkit/pkg/platform/dhcp"
	"github.com/sakai135/wsl-vpnkit/pkg/platform/link"
	"github.com/sakai135/wsl-vpnkit/pkg/platform/tap"
	"github.com/sakai135/wsl-vpnkit/pkg/transport"
	log "github.com/sirupsen/logrus"
)

func main() {
	f, err := parseVMFlags()
	if err != nil {
		log.Fatal(err)
	}

	if err := run(f); err != nil {
		log.Fatal(err)
		if err == io.EOF {
			os.Exit(1)
		}
	}
}

func parseProxyOptions(f *VMFlags) []string {
	options := []string{
		"-subnet", f.Subnet,
		"-gateway-ip", f.GatewayIP,
		"-host-ip", f.HostIP,
		"-vm-ip", f.VMIP,
		"-mtu", strconv.Itoa(f.ProxyMTU),
	}
	if f.Debug {
		options = append(options, "-debug")
	}
	return options
}

func run(f *VMFlags) error {
	conn, err := transport.Dial(f.Endpoint, parseProxyOptions(f)[:]...)
	if err != nil {
		return errors.Wrap(err, "cannot connect to host")
	}
	defer conn.Close()

	tap, err := tap.New(f.Iface, conn, f.MTU, f.Debug)
	if err != nil {
		return errors.Wrap(err, "cannot create tap device")
	}
	defer tap.Close()

	if err := link.LinkUp(f.Iface, f.MAC); err != nil {
		return errors.Wrap(err, "cannot set mac address")
	}

	errCh := make(chan error, 1)
	tap.Start(errCh)
	go func() {
		if err := dhcp.RequestDHCP(f.Iface); err != nil {
			errCh <- errors.Wrap(err, "dhcp error")
		}
	}()
	return <-errCh
}
