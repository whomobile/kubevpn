package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type datagramPacket struct {
	DataLength uint16 // [2]byte
	Data       []byte // []byte
}

func (addr *datagramPacket) String() string {
	if addr == nil {
		return ""
	}
	return fmt.Sprintf("DataLength: %d, Data: %v\n", addr.DataLength, addr.Data)
}

func newDatagramPacket(data []byte) (r *datagramPacket) {
	return &datagramPacket{
		DataLength: uint16(len(data)),
		Data:       data,
	}
}

func (addr *datagramPacket) Addr() net.Addr {
	return Server8422
}

func readDatagramPacket(r io.Reader, b []byte) (*datagramPacket, error) {
	_, err := io.ReadFull(r, b[:2])
	if err != nil {
		return nil, err
	}
	dataLength := binary.BigEndian.Uint16(b[:2])
	if _, err = io.ReadFull(r, b[:dataLength]); err != nil && (err != io.ErrUnexpectedEOF || err != io.EOF) {
		return nil, err
	}
	return &datagramPacket{DataLength: dataLength, Data: b[:dataLength]}, nil
}

func (addr *datagramPacket) Write(w io.Writer) error {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b[:], uint16(len(addr.Data)))
	if _, err := w.Write(b); err != nil {
		return err
	}
	if _, err := w.Write(addr.Data); err != nil {
		return err
	}
	return nil
}
