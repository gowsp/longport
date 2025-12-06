package longport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	controlv1 "github.com/longportapp/openapi-protobufs/gen/go/control"
	"google.golang.org/protobuf/proto"
)

type rpc func(body *Body, err error)

type ConnType string

const (
	QUOTE ConnType = "quote"
	TRADE ConnType = "trade"
)

type websocket struct {
	start sync.Once

	url  string
	api  Api
	req  sync.Map
	conn net.Conn
	rid  atomic.Uint32

	handler func(event *Event) error
}

func (w *websocket) Rpc(cmd byte, req proto.Message, rsp proto.Message) error {
	w.start.Do(func() {
		w.connect()
	})
	return w.rpc(cmd, req, rsp)
}
func (w *websocket) Subscribe(handler func(body *Event) error) {
	w.handler = handler
}
func (w *websocket) rpc(cmd byte, req proto.Message, rsp proto.Message) error {
	id := w.rid.Add(1)
	auth, err := wsRequst(id, cmd, req)
	if err != nil {
		return err
	}
	res := make(chan error, 1)
	var rpc rpc = func(body *Body, err error) {
		if err == nil {
			err = body.UnmarshalProto(rsp)
		}
		res <- err
	}
	w.req.Store(id, rpc)
	err = wsutil.WriteClientBinary(w.conn, auth)
	if err != nil {
		return err
	}
	result := <-res
	return result
}

func (w *websocket) reconnect() (*controlv1.AuthResponse, error) {
	rsp := new(controlv1.AuthResponse)
	if time.Now().UnixMilli() >= rsp.Expires {
		return nil, errors.New("session expires")
	}
	req := controlv1.ReconnectRequest{SessionId: rsp.SessionId}
	return rsp, w.rpc(byte(controlv1.Command_CMD_RECONNECT), &req, rsp)
}

type tokenRsp struct {
	Otp string `json:"otp"`
}

func (w *websocket) getToken() (string, error) {
	rsp := new(tokenRsp)
	if err := w.api.Get("/v1/socket/token", rsp, nil); err != nil {
		return "", err
	}
	return rsp.Otp, nil
}
func (w *websocket) auth() (*controlv1.AuthResponse, error) {
	token, err := w.getToken()
	if err != nil {
		return nil, err
	}
	rsp := new(controlv1.AuthResponse)
	req := controlv1.AuthRequest{Token: token}
	return rsp, w.rpc(byte(controlv1.Command_CMD_AUTH), &req, rsp)
}
func (w *websocket) login() error {
	_, err := w.reconnect()
	if err != nil {
		_, err = w.auth()
	}
	return err
}

func (w *websocket) connect() (err error) {
	log.Println("connect:", w.url)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	w.conn, _, _, err = ws.Dial(ctx, w.url)
	if err != nil {
		log.Println("error", err)
		return fmt.Errorf("failed to connect to %s: %w", w.url, err)
	}

	go w.serve(w.connect)
	return w.login()
}
func (w *websocket) serve(reload func() error) {
	for {
		d, err := wsutil.ReadServerBinary(w.conn)
		if err != nil {
			log.Println("read server data error", err)
			break
		}
		p := new(packet)
		if err = p.UnmarshalBinary(d); err != nil {
			log.Println(err)
			continue
		}
		w.hanlde(p)
	}
	reload()
}

func (w *websocket) hanlde(p *packet) {
	switch p.header.Type {
	case 2:
		rs := new(response)
		rs.UnmarshalBinary(p.data)
		val, ok := w.req.LoadAndDelete(rs.id)
		if !ok {
			return
		}
		if rs.status == 0 {
			rs.body.gzip = p.header.Gzip
			val.(rpc)(rs.body, nil)
			return
		}
		rsp := new(controlv1.Error)
		rs.body.UnmarshalProto(rsp)
		val.(rpc)(rs.body, errors.New(rsp.Msg))
	case 3:
		event := new(Event)
		if err := event.UnmarshalBinary(p.data); err != nil {
			log.Println(err)
		} else {
			w.handler(event)
		}
	}
}

func (w *websocket) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}
