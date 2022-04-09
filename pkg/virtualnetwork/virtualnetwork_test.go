package virtualnetwork

import (
	"net"
	"net/http"
	"testing"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
)

func TestNew(t *testing.T) {
	config := &types.Configuration{
		Debug:             false,
		CaptureFile:       "",
		MTU:               1500,
		Subnet:            "192.168.127.0/24",
		GatewayIP:         "192.168.127.1",
		GatewayMacAddress: "5a:94:ef:e4:0c:dd",
		DHCPStaticLeases: map[string]string{
			"192.168.127.2": "5a:94:ef:e4:0c:ee",
		},
		DNS:              []types.Zone{},
		DNSSearchDomains: nil,
		Forwards:         map[string]string{},
		NAT: map[string]string{
			"192.168.127.254": "127.0.0.1",
		},
		GatewayVirtualIPs:      []string{"192.168.127.254"},
		VpnKitUUIDMacAddresses: map[string]string{},
		Protocol:               types.HyperKitProtocol,
	}

	vn, err := New(config)
	if err != nil {
		t.Error(err)
	}

	ln, err := vn.ListenTCP(net.ParseIP("192.168.127.1"), 9090)
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hello"))
	})
	go func() {
		_ = http.Serve(ln, mux)
	}()
}
