package control

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gowsp/longport/config"
	"google.golang.org/protobuf/proto"
)

type rpc func(body *Body, err error)

// 推送消息处理
type PushEvent func(body *Push) error

// 新建长连接
func NewConn(api *Invoker, ct config.ConnType) *Conn {
	return &Conn{api: api, ct: ct}
}

type Conn struct {
	init sync.Once

	ct  config.ConnType
	api *Invoker
	req sync.Map
	wss net.Conn
	rid atomic.Uint32

	Event PushEvent
}

func (w *Conn) Rpc(cmd byte, req proto.Message, rsp proto.Message) error {
	w.init.Do(func() {
		w.connect()
	})
	return w.rpc(cmd, req, rsp)
}
func (w *Conn) rpc(cmd byte, req proto.Message, rsp proto.Message) error {
	id := w.rid.Add(1)
	auth, err := NewRequst(id, cmd, req)
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
	err = wsutil.WriteClientBinary(w.wss, auth)
	if err != nil {
		return err
	}
	result := <-res
	return result
}

func (w *Conn) reconnect() (*AuthResponse, error) {
	rsp := new(AuthResponse)
	if time.Now().UnixMilli() >= rsp.Expires {
		return nil, errors.New("session expires")

	}
	req := ReconnectRequest{SessionId: rsp.SessionId}
	return rsp, w.rpc(byte(Command_CMD_RECONNECT), &req, rsp)
}

type tokenRsp struct {
	Otp string `json:"otp"`
}

func (a *Conn) getToken() (string, error) {
	rsp := new(tokenRsp)
	if err := a.api.Get("/v1/socket/token", rsp, nil); err != nil {
		return "", err
	}
	return rsp.Otp, nil
}
func (w *Conn) auth() (*AuthResponse, error) {
	token, err := w.getToken()
	if err != nil {
		return nil, err
	}
	rsp := new(AuthResponse)
	req := AuthRequest{Token: token}
	return rsp, w.rpc(byte(Command_CMD_AUTH), &req, rsp)
}
func (w *Conn) login() error {
	_, err := w.reconnect()
	if err != nil {
		_, err = w.auth()
	}
	return err
}

func (w *Conn) connect() (err error) {
	url := w.api.config.Conn(w.ct)
	log.Println("connect:", url)
	w.wss, _, _, err = ws.Dial(context.Background(), url)
	if err != nil {
		log.Println("error", err)
		return err
	}
	go w.serve(w.connect)
	return w.login()
}
func (w *Conn) serve(reload func() error) {
	for {
		d, err := wsutil.ReadServerBinary(w.wss)
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

func (w *Conn) hanlde(p *packet) {
	switch p.header.Type {
	case 2:
		rs := new(reponse)
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
		rsp := new(Error)
		rs.body.UnmarshalProto(rsp)
		val.(rpc)(rs.body, errors.New(rsp.Msg))
	case 3:
		push := new(Push)
		if err := push.UnmarshalBinary(p.data); err != nil {
			log.Println(err)
		} else {
			w.Event(push)
		}
	}
}
