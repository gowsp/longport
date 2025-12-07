package longport

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"

	tradev1 "github.com/longportapp/openapi-protobufs/gen/go/trade"
	"github.com/shopspring/decimal"
)

// 交易长连接
type TradeConn interface {
	// 订阅
	Subscribe(func(*OrderEvent)) error
	// 取消订阅
	Unsubscribe() error
}

type tradeConn struct {
	*websocket
}
type OrderEvent struct {
	Event string `json:"event,omitempty"`
	Data  struct {
		AccountNo string `json:"account_no,omitempty"`
		LastShare string `json:"last_share,omitempty"`
		LastPrice string `json:"last_price,omitempty"`

		SubmittedPrice    decimal.Decimal `json:"submitted_price,omitempty"`
		SubmittedQuantity decimal.Decimal `json:"submitted_quantity,omitempty"`
		*CommonOrder
	} `json:"data"`
}

func (t *tradeConn) decode(n *tradev1.Notification) (*OrderEvent, error) {
	event := new(OrderEvent)
	switch n.ContentType {
	case tradev1.ContentType_CONTENT_JSON:
		val, err := base64.RawStdEncoding.DecodeString(string(n.Data))
		if err != nil {
			return nil, err
		}
		log.Println("recive order event", string(val))
		return event, json.Unmarshal(val, event)
	case tradev1.ContentType_CONTENT_UNDEFINED:
		log.Println("recive order event", string(n.Data))
		return event, json.Unmarshal(n.Data, event)
	}
	return event, errors.New("not found")
}
func (t *tradeConn) Subscribe(handler func(*OrderEvent)) error {
	t.websocket.Subscribe(func(event *event) error {
		switch event.Cmd {
		case byte(tradev1.Command_CMD_NOTIFY):
			notify := new(tradev1.Notification)
			if err := event.UnmarshalProto(notify); err != nil {
				return err
			}
			u, err := t.decode(notify)
			if err != nil {
				log.Println(err)
				return err
			}
			handler(u)
		}
		return nil
	})
	rsp := new(tradev1.SubResponse)
	return t.Rpc(byte(tradev1.Command_CMD_SUB), &tradev1.Sub{Topics: []string{"private"}}, rsp)
}

func (t *tradeConn) Unsubscribe() error {
	rsp := new(tradev1.UnsubResponse)
	return t.Rpc(byte(tradev1.Command_CMD_UNSUB), &tradev1.Unsub{Topics: []string{"private"}}, rsp)
}
