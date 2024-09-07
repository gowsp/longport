package control

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/proto"
)

type MsgType uint32

const (
	REQUEST  MsgType = 1
	RESOPNSE MsgType = 2
	PUSH     MsgType = 3
)

type header struct {
	Type    uint8
	Verify  uint8
	Gzip    uint8
	Reserve uint8
}

func (h *header) MarshalBinary() (data []byte, err error) {
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

func NewRequst(id uint32, cmd byte, body proto.Message) ([]byte, error) {
	var err error
	var data []byte
	if body != nil {
		data, err = proto.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req := request{cmd: cmd, id: id, timeout: 60000, body: &Body{body: data}}
	if data, err = req.MarshalBinary(); err != nil {
		return nil, err

	}
	p := &packet{header: &header{Type: uint8(REQUEST)}, data: data}
	return p.MarshalBinary()
}

// 网络包
type packet struct {
	header *header
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
	h := new(header)
	b := data[0]
	h.Type = 0xf & b
	h.Verify = b >> 4 & 0x1
	h.Gzip = b >> 5 & 0x1
	h.Reserve = b >> 6 & 0x3
	r.header = h
	r.data = data[1:]
	return nil
}

// 指令请求
type request struct {
	id      uint32
	cmd     byte
	timeout uint16
	body    *Body
}

func (r *request) MarshalBinary() (data []byte, err error) {
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

// 指令响应
type reponse struct {
	id     uint32
	cmd    byte
	status byte
	body   *Body
}

func (r *reponse) UnmarshalBinary(data []byte) error {
	r.cmd = data[0]
	r.id = binary.BigEndian.Uint32(data[1:])
	r.status = data[5]
	r.body = new(Body)
	return r.body.UnmarshalBinary(data[5:])
}

// 来自长桥的推送消息
type Push struct {
	Cmd  byte  // 推送指令
	body *Body // 推送消息体
}

// 二进制消息体转Proto
func (p *Push) UnmarshalProto(msg proto.Message) error {
	return p.body.UnmarshalProto(msg)
}

// 解码推送消息体
func (p *Push) UnmarshalBinary(data []byte) error {
	p.Cmd = data[0]
	p.body = new(Body)
	return p.body.UnmarshalBinary(data)
}
