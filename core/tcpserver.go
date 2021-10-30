package core

import (
	"bytes"
	"context"
	"fmt"
	"github.com/wencaiwulue/kubevpn/util"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type fakeUDPTunConnector struct {
}

// UDPOverTCPTunnelConnector creates a connector for UDP-over-TCP
func UDPOverTCPTunnelConnector() Connector {
	return &fakeUDPTunConnector{}
}

func (c *fakeUDPTunConnector) ConnectContext(_ context.Context, conn net.Conn, network, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return nil, fmt.Errorf("%s unsupported", network)
	}
	_ = conn.SetDeadline(time.Now().Add(util.ConnectTimeout))
	defer conn.SetDeadline(time.Time{})

	targetAddr, _ := net.ResolveUDPAddr("udp", address)
	return newFakeUDPTunnelConnOverTCP(conn, targetAddr)
}

type fakeUdpHandler struct {
}

// TCPHandler creates a server Handler
func TCPHandler() Handler {
	h := &fakeUdpHandler{}
	h.Init()

	return h
}

func (h *fakeUdpHandler) Init(...HandlerOption) {
}

func (h *fakeUdpHandler) Handle(conn net.Conn) {
	defer conn.Close()
	if util.Debug {
		log.Debugf("[tcpserver] %s -> %s\n", conn.RemoteAddr(), conn.LocalAddr())
	}
	h.handleUDPTunnel(conn)
}

func (h *fakeUdpHandler) transportUDP(relay, peer net.PacketConn) (err error) {
	errc := make(chan error, 2)

	var clientAddr net.Addr

	go func() {
		b := util.MPool.Get().([]byte)
		defer util.MPool.Put(b)

		for {
			n, laddr, err := relay.ReadFrom(b)
			if err != nil {
				errc <- err
				return
			}
			if clientAddr == nil {
				clientAddr = laddr
			}
			dgram, err := ReadDatagramPacket(bytes.NewReader(b[:n]))
			if err != nil {
				log.Errorln(err)
				errc <- err
				return
			}

			raddr, err := net.ResolveUDPAddr("udp", dgram.Addr())
			if err != nil {
				log.Debugf("[tcpserver-udp] addr error, addr: %s, err: %v", dgram.Addr(), err)
				continue // drop silently
			}
			if _, err := peer.WriteTo(dgram.Data, raddr); err != nil {
				errc <- err
				return
			}
			if util.Debug {
				log.Debugf("[tcpserver-udp] %s >>> %s length: %d", relay.LocalAddr(), raddr, len(dgram.Data))
			}
		}
	}()

	go func() {
		b := util.MPool.Get().([]byte)
		defer util.MPool.Put(b)

		for {
			n, raddr, err := peer.ReadFrom(b)
			if err != nil {
				errc <- err
				return
			}
			if clientAddr == nil {
				continue
			}
			buf := bytes.Buffer{}
			dgram := NewDatagramPacket(raddr, b[:n])
			_ = dgram.Write(&buf)
			if _, err := relay.WriteTo(buf.Bytes(), clientAddr); err != nil {
				errc <- err
				return
			}
			if util.Debug {
				log.Debugf("[tcpserver-udp] %s <<< %s length: %d", relay.LocalAddr(), raddr, len(dgram.Data))
			}
		}
	}()

	return <-errc
}

func (h *fakeUdpHandler) handleUDPTunnel(conn net.Conn) {
	// serve tunnel udp, tunnel <-> remote, handle tunnel udp request
	bindAddr, _ := net.ResolveUDPAddr("udp", ":0")
	uc, err := net.ListenUDP("udp", bindAddr)
	if err != nil {
		log.Debugf("[tcpserver] udp-tun %s -> %s : %s", conn.RemoteAddr(), bindAddr, err)
		return
	}
	defer uc.Close()
	if util.Debug {
		log.Debugf("[tcpserver] udp-tun %s <- %s\n", conn.RemoteAddr(), uc.LocalAddr())
	}
	log.Debugf("[tcpserver] udp-tun %s <-> %s", conn.RemoteAddr(), uc.LocalAddr())
	_ = h.tunnelServerUDP(conn, uc)
	log.Debugf("[tcpserver] udp-tun %s >-< %s", conn.RemoteAddr(), uc.LocalAddr())
	return
}

func (h *fakeUdpHandler) tunnelServerUDP(cc net.Conn, pc net.PacketConn) (err error) {
	errc := make(chan error, 2)

	go func() {
		b := util.MPool.Get().([]byte)
		defer util.MPool.Put(b)

		for {
			n, addr, err := pc.ReadFrom(b)
			if err != nil {
				log.Debugf("[udp-tun] %s : %s", cc.RemoteAddr(), err)
				errc <- err
				return
			}

			// pipe from peer to tunnel
			dgram := NewDatagramPacket(addr, b[:n])
			if err := dgram.Write(cc); err != nil {
				log.Debugf("[tcpserver] udp-tun %s <- %s : %s", cc.RemoteAddr(), dgram.Addr(), err)
				errc <- err
				return
			}
			if util.Debug {
				log.Debugf("[tcpserver] udp-tun %s <<< %s length: %d", cc.RemoteAddr(), dgram.Addr(), len(dgram.Data))
			}
		}
	}()

	go func() {
		for {
			dgram, err := ReadDatagramPacket(cc)
			if err != nil {
				log.Debugf("[udp-tun] %s -> 0 : %v", cc.RemoteAddr(), err)
				errc <- err
				return
			}

			// pipe from tunnel to peer
			addr, err := net.ResolveUDPAddr("udp", dgram.Addr())
			if err != nil {
				log.Debugf("[tcpserver-udp] addr error, addr: %s, err: %v", dgram.Addr(), err)
				continue // drop silently
			}
			if _, err := pc.WriteTo(dgram.Data, addr); err != nil {
				log.Debugf("[tcpserver] udp-tun %s -> %s : %s", cc.RemoteAddr(), addr, err)
				errc <- err
				return
			}
			if util.Debug {
				log.Debugf("[tcpserver] udp-tun %s >>> %s length: %d", cc.RemoteAddr(), addr, len(dgram.Data))
			}
		}
	}()

	return <-errc
}

// fake udp connect over tcp
type fakeUDPTunnelConn struct {
	// tcp connection
	net.Conn
	targetAddr net.Addr
}

func newFakeUDPTunnelConnOverTCP(conn net.Conn, targetAddr net.Addr) (net.Conn, error) {
	return &fakeUDPTunnelConn{
		Conn:       conn,
		targetAddr: targetAddr,
	}, nil
}

func (c *fakeUDPTunnelConn) Read(b []byte) (n int, err error) {
	n, _, err = c.ReadFrom(b)
	return
}

func (c *fakeUDPTunnelConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	dgram, err := ReadDatagramPacket(c.Conn)
	if err != nil {
		log.Errorln(err)
		return
	}
	n = copy(b, dgram.Data)
	addr, err = net.ResolveUDPAddr("udp", dgram.Addr())
	if err != nil {
		log.Debugf("[tcpserver-udp] addr error, addr: %s, err: %v", dgram.Addr(), err)
	}
	return
}

func (c *fakeUDPTunnelConn) Write(b []byte) (n int, err error) {
	return c.WriteTo(b, c.targetAddr)
}

func (c *fakeUDPTunnelConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	dgram := NewDatagramPacket(addr, b)
	if err = dgram.Write(c.Conn); err != nil {
		return
	}
	return len(b), nil
}