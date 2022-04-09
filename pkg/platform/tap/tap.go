package tap

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type Tap interface {
	Start(errCh chan error)
	Close() error
}

type tap struct {
	tap   *water.Interface
	name  string
	conn  net.Conn
	mtu   int
	debug bool
}

func New(name string, conn net.Conn, mtu int, debug bool) (Tap, error) {
	ifce, err := water.New(water.Config{
		DeviceType: water.TAP,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: name,
		},
	})
	if err != nil {
		return nil, err
	}
	return &tap{
		tap:   ifce,
		name:  name,
		conn:  conn,
		mtu:   mtu,
		debug: debug,
	}, nil
}

func (t *tap) Start(errCh chan error) {
	go t.tx(errCh)
	go t.rx(errCh)
}

func (t *tap) Close() error {
	return t.tap.Close()
}

func (t *tap) rx(errCh chan error) {
	log.Info("waiting for packets...")
	var frame ethernet.Frame
	for {
		frame.Resize(t.mtu)
		n, err := t.tap.Read([]byte(frame))
		if err != nil {
			errCh <- errors.Wrap(err, "cannot read packet from tap")
			return
		}
		frame = frame[:n]

		if t.debug {
			packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
			log.Info(packet.String())
		}

		size := make([]byte, 2)
		binary.LittleEndian.PutUint16(size, uint16(n))

		if _, err := t.conn.Write(size); err != nil {
			errCh <- errors.Wrap(err, "cannot write size to socket")
			return
		}
		if _, err := t.conn.Write(frame); err != nil {
			errCh <- errors.Wrap(err, "cannot write packet to socket")
			return
		}
	}
}

func (t *tap) tx(errCh chan error) {
	sizeBuf := make([]byte, 2)
	buf := make([]byte, t.mtu+header.EthernetMinimumSize)

	for {
		n, err := io.ReadFull(t.conn, sizeBuf)
		if err != nil {
			errCh <- errors.Wrap(err, "cannot read size from socket")
			return
		}
		if n != 2 {
			errCh <- fmt.Errorf("unexpected size %d", n)
			return
		}
		size := int(binary.LittleEndian.Uint16(sizeBuf[0:2]))

		n, err = io.ReadFull(t.conn, buf[:size])
		if err != nil {
			errCh <- errors.Wrap(err, "cannot read payload from socket")
			return
		}
		if n == 0 || n != size {
			errCh <- fmt.Errorf("unexpected size %d != %d", n, size)
			return
		}

		if t.debug {
			packet := gopacket.NewPacket(buf[:size], layers.LayerTypeEthernet, gopacket.Default)
			log.Info(packet.String())
		}

		if _, err := t.tap.Write(buf[:size]); err != nil {
			errCh <- errors.Wrap(err, "cannot write packet to tap")
			return
		}
	}
}
