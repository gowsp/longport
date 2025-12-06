// Package longport provides functionality for communicating with the LongPort API
// through WebSocket connections with protobuf message encoding.
package longport

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/proto"
)

// MsgType represents the type of message being sent or received
type MsgType uint32

const (
	REQUEST  MsgType = 1 // Outgoing request message
	RESPONSE MsgType = 2 // Incoming response message
	PUSH     MsgType = 3 // Server-initiated push message
)

// Header represents the message header containing metadata
type Header struct {
	Type    uint8 // Message type
	Verify  uint8 // Verification flag
	Gzip    uint8 // Compression flag
	Reserve uint8 // Reserved field
}

func (h *Header) MarshalBinary() (data []byte, err error) {
	b := (h.Type & 0xf) | ((h.Verify & 0x1) << 4) | ((h.Gzip & 0x1) << 5) | ((h.Reserve & 0x3) << 6)
	data = append(data, b)
	return
}

type Body struct {
	gzip uint8
	body []byte
}

func (b *Body) UnmarshalProto(msg proto.Message) error {
	if b.gzip == 0 {
		return proto.Unmarshal(b.body, msg)
	}
	r, err := gzip.NewReader(bytes.NewBuffer(b.body))
	if err != nil {
		return err
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, msg)
}
func (b *Body) MarshalBinary() (data []byte, err error) {
	len := len(b.body)
	data = binary.BigEndian.AppendUint32(data, uint32(len))
	data = append(data, b.body...)
	return
}
func (b *Body) UnmarshalBinary(data []byte) error {
	data[0] = 0
	len := binary.BigEndian.Uint32(data)
	b.body = data[4 : 4+len]
	return nil
}

func wsRequst(id uint32, cmd byte, body proto.Message) ([]byte, error) {
	var err error
	var data []byte
	if body != nil {
		data, err = proto.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req := Request{cmd: cmd, id: id, timeout: 60000, body: &Body{body: data}}
	if data, err = req.MarshalBinary(); err != nil {
		return nil, err
	}
	p := &packet{header: &Header{Type: uint8(REQUEST)}, data: data}
	return p.MarshalBinary()
}

type packet struct {
	header *Header
	data   []byte
}

func (r *packet) MarshalBinary() (data []byte, err error) {
	v, err := r.header.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data = append(data, v...)
	data = append(data, r.data...)
	return
}

func (r *packet) UnmarshalBinary(data []byte) error {
	h := new(Header)
	b := data[0]
	h.Type = 0xf & b
	h.Verify = b >> 4 & 0x1
	h.Gzip = b >> 5 & 0x1
	h.Reserve = b >> 6 & 0x3
	r.header = h
	r.data = data[1:]
	return nil
}

type Request struct {
	id      uint32
	cmd     byte
	timeout uint16
	body    *Body
}

func (r *Request) MarshalBinary() (data []byte, err error) {
	data = append(data, r.cmd)
	data = binary.BigEndian.AppendUint32(data, r.id)
	data = binary.BigEndian.AppendUint16(data, r.timeout)
	body, err := r.body.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data = append(data, body[1:]...)
	return
}

type response struct {
	id     uint32
	cmd    byte
	status byte
	body   *Body
}

func (r *response) UnmarshalBinary(data []byte) error {
	r.cmd = data[0]
	r.id = binary.BigEndian.Uint32(data[1:])
	r.status = data[5]
	r.body = new(Body)
	return r.body.UnmarshalBinary(data[5:])
}

type Event struct {
	Cmd  byte
	body *Body
}

func (p *Event) UnmarshalProto(msg proto.Message) error {
	return p.body.UnmarshalProto(msg)
}
func (p *Event) UnmarshalBinary(data []byte) error {
	p.Cmd = data[0]
	p.body = new(Body)
	return p.body.UnmarshalBinary(data)
}
