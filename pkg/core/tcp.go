package core

import (
	"context"
	"errors"
	"net"

	"github.com/wencaiwulue/kubevpn/pkg/config"
)

type tcpTransporter struct{}

func TCPTransporter() Transporter {
	return &tcpTransporter{}
}

func (tr *tcpTransporter) Dial(ctx context.Context, addr string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: config.DialTimeout}
	return dialer.DialContext(ctx, "tcp", addr)
}

func TCPListener(addr string) (net.Listener, error) {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		err = errors.New("net.ResolveTCPAddr(\"tcp\", addr): " + err.Error())
		return nil, err
	}
	ln, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		err = errors.New("net.ListenTCP(\"tcp\", laddr): " + err.Error())
		return nil, err
	}
	return &tcpKeepAliveListener{ln}, nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln *tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	conn, err := ln.AcceptTCP()
	if err != nil {
		err = errors.New("ln.AcceptTCP(): " + err.Error())
		return
	}
	err = conn.SetKeepAlive(true)
	if err != nil {
		err = errors.New("conn.SetKeepAlive(true): " + err.Error())
		return nil, err
	}
	err = conn.SetKeepAlivePeriod(config.KeepAliveTime)
	if err != nil {
		err = errors.New("conn.SetKeepAlivePeriod(config.KeepAliveTime): " + err.Error())
		return nil, err
	}
	err = conn.SetNoDelay(true)
	if err != nil {
		err = errors.New("conn.SetNoDelay(true): " + err.Error())
		return nil, err
	}
	return conn, nil
}
