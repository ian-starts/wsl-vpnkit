package link

import (
	"errors"
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func LinkUp(iface string, mac string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}
	if mac == "" {
		return netlink.LinkSetUp(link)
	}
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}
	if err := netlink.LinkSetHardwareAddr(link, hw); err != nil {
		return err
	}
	return netlink.LinkSetUp(link)
}

func VerifyInterface(iface string) error {
	links, err := netlink.LinkList()
	if err != nil {
		return err
	}
	for _, link := range links {
		if iface == link.Attrs().Name {
			return errors.New(fmt.Sprintf("interface %s prevented this program to run", link.Attrs().Name))
		}
	}
	return nil
}
