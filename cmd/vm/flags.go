package main

import (
	"flag"
	"net"
	"os"

	"github.com/pkg/errors"
	"github.com/sakai135/wsl-vpnkit/pkg/platform/link"
)

const (
	vmMacAddress = "5a:94:ef:e4:0c:ee"
)

type VMFlags struct {
	Endpoint string
	Iface    string
	Debug    bool
	MTU      int

	Subnet    string
	GatewayIP string
	HostIP    string
	VMIP      string
	ProxyMTU  int
	MAC       string
}

func parseVMFlags() (*VMFlags, error) {
	f := &VMFlags{}

	flag.StringVar(&f.Endpoint, "path", "gvproxy.exe", "path to gvproxy.exe")
	flag.StringVar(&f.Iface, "iface", "tap0", "tap interface name")
	flag.BoolVar(&f.Debug, "debug", false, "debug")
	flag.IntVar(&f.MTU, "mtu", 4000, "mtu")

	flag.StringVar(&f.Subnet, "subnet", "192.168.127.0/24", "Set the subnet")
	flag.StringVar(&f.GatewayIP, "gateway-ip", "192.168.127.1", "Set the IP for the gateway")
	flag.StringVar(&f.HostIP, "host-ip", "192.168.127.254", "Set the IP for accessing the host from the WSL 2 VM")
	flag.StringVar(&f.VMIP, "vm-ip", "192.168.127.2", "Set the IP for the WSL 2 VM")
	flag.IntVar(&f.ProxyMTU, "proxy-mtu", 1500, "Set the MTU for the proxy")

	flag.Parse()

	if _, err := os.Stat(f.Endpoint); err != nil {
		return nil, errors.Wrapf(err, "error verifying path %s", f.Endpoint)
	}
	if net.ParseIP(f.GatewayIP) == nil {
		return nil, errors.New("invalid gateway-ip")
	}
	if net.ParseIP(f.HostIP) == nil {
		return nil, errors.New("invalid host-ip")
	}
	if net.ParseIP(f.VMIP) == nil {
		return nil, errors.New("invalid vm-ip")
	}
	if _, _, err := net.ParseCIDR(f.Subnet); err != nil {
		return nil, errors.Wrap(err, "invalid subnet")
	}
	if err := link.VerifyInterface(f.Iface); err != nil {
		return nil, errors.Wrap(err, "invalid iface")
	}

	f.MAC = vmMacAddress

	return f, nil
}
