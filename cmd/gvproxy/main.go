package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	meikgDNS "github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/sakai135/wsl-vpnkit/pkg/services/dns"
	"github.com/sakai135/wsl-vpnkit/pkg/transport"
	"github.com/sakai135/wsl-vpnkit/pkg/virtualnetwork"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	exitCode int
)

func main() {
	log.SetOutput(os.Stderr)

	config, err := parseProxyFlags()
	if err != nil {
		log.Error(err)
		exitCode = 1
		return
	}

	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Make this the last defer statement in the stack
	defer os.Exit(exitCode)

	groupErrs, ctx := errgroup.WithContext(ctx)
	// Setup signal channel for catching user signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	vn, err := virtualnetwork.New(config)
	if err != nil {
		log.Error(err)
		exitCode = 1
		return
	}
	dnsServer := dns.NewServer(vn, dns.NewHandler(config.DNS, &net.Resolver{PreferGo: false}), meikgDNS.ActivateAndServe)
	gatewayIP := net.ParseIP(config.GatewayIP)
	groupErrs.Go(func() error {
		go func() {
			err := dnsServer.StartTCP(gatewayIP, 53)
			if err != nil {
				log.Error(err)
			}
		}()
		go func() {
			err := dnsServer.StartUDP(gatewayIP, 53)
			if err != nil {
				log.Error(err)
			}
		}()
		conn := transport.GetStdioConn()
		return vn.Accept(ctx, conn)
	})

	// Wait for something to happen
	groupErrs.Go(func() error {
		select {
		// Catch signals so exits are graceful and defers can run
		case <-sigChan:
			cancel()
			return errors.New("signal caught")
		case <-ctx.Done():
			return nil
		}
	})
	// Wait for all of the go funcs to finish up
	if err := groupErrs.Wait(); err != nil {
		log.Error(err)
		exitCode = 1
	}
}
