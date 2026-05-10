package longport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	controlv1 "github.com/longportapp/openapi-protobufs/gen/go/control"
	"google.golang.org/protobuf/proto"
)

type rpc func(body *body, err error)

const defaultRPCTimeout = 15 * time.Second

type websocket struct {
	start sync.Once

	url  string
	api  Api
	req  sync.Map
	conn net.Conn
	rid  atomic.Uint32

	handler func(event *event) error
}

func (w *websocket) Rpc(cmd byte, req proto.Message, rsp proto.Message) error {
	var err error
	w.start.Do(func() {
		err = w.connect()
	})
	if err != nil {
		return err
	}
	if w.conn == nil {
		return errors.New("websocket not connected")
	}
	return w.rpc(cmd, req, rsp)
}
func (w *websocket) Subscribe(handler func(body *event) error) {
	w.handler = handler
}
func (w *websocket) rpc(cmd byte, req proto.Message, rsp proto.Message) error {
	id := w.rid.Add(1)
	auth, err := wsRequst(id, cmd, req)
	if err != nil {
		return err
	}
	res := make(chan error, 1)
	var rpc rpc = func(body *body, err error) {
		if err == nil {
			err = body.UnmarshalProto(rsp)
		}
		res <- err
	}
	w.req.Store(id, rpc)
	err = wsutil.WriteClientBinary(w.conn, auth)
	if err != nil {
		w.req.Delete(id)
		return err
	}
	timer := time.NewTimer(rpcTimeout())
	defer timer.Stop()
	select {
	case result := <-res:
		return result
	case <-timer.C:
		w.req.Delete(id)
		return fmt.Errorf("websocket rpc timeout: cmd=%d id=%d", cmd, id)
	}
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
	var readErr error
	for {
		d, err := wsutil.ReadServerBinary(w.conn)
		if err != nil {
			log.Println("read server data error", err)
			readErr = err
			break
		}
		p := new(packet)
		if err = p.UnmarshalBinary(d); err != nil {
			log.Println(err)
			continue
		}
		w.hanlde(p)
	}
	w.failPending(fmt.Errorf("websocket disconnected: %w", readErr))
	if err := reload(); err != nil {
		log.Println("reload websocket error", err)
	}
}

func (w *websocket) failPending(err error) {
	w.req.Range(func(key, value any) bool {
		if id, ok := key.(uint32); ok {
			w.req.Delete(id)
		} else {
			w.req.Delete(key)
		}
		value.(rpc)(nil, err)
		return true
	})
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
		event := new(event)
		if err := event.UnmarshalBinary(p.data); err != nil {
			log.Println(err)
		} else if w.handler != nil {
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

func rpcTimeout() time.Duration {
	val := os.Getenv("LONGPORT_RPC_TIMEOUT")
	if val == "" {
		return defaultRPCTimeout
	}
	timeout, err := time.ParseDuration(val)
	if err == nil {
		return timeout
	}
	seconds, err := strconv.Atoi(val)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}
	return defaultRPCTimeout
}
