package trade

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gowsp/longport/config"
	"github.com/gowsp/longport/control"
)

type Event func(*OrderEvent)

type OrderEvent struct {
	Event string `json:"event"`
	Data  struct {
		AccountNo string `json:"account_no"`
		LastShare string `json:"last_share"`
		LastPrice string `json:"last_price"`
		*CommonOrder
	} `json:"data"`
}

func (e *OrderEvent) String() string {
	buff := new(bytes.Buffer)
	fmt.Fprintf(buff, "%s\n%s %s\n", e.Data.AccountNo, e.Data.Symbol, e.Data.StockName)
	fmt.Fprintf(buff, "%s 成交价 %s, 成交数量 %s\n", e.Data.Side, e.Data.ExecutedPrice, e.Data.ExecutedQuantity)
	return buff.String()
}

func (n *Notification) decode() (*OrderEvent, error) {
	event := new(OrderEvent)
	switch n.ContentType {
	case ContentType_CONTENT_JSON:
		val, err := base64.RawStdEncoding.DecodeString(string(n.Data))
		if err != nil {
			return nil, err
		}
		return event, json.Unmarshal(val, event)
	case ContentType_CONTENT_UNDEFINED:
		return event, json.Unmarshal(n.Data, event)
	}
	return event, errors.New("not found")
}
func (t *trader) TredeEvent(event Event) {
	conn := control.NewConn(t.api, config.TRADE)
	conn.Event = func(push *control.Push) error {
		switch push.Cmd {
		case byte(Command_CMD_NOTIFY):
			notify := new(Notification)
			err := push.UnmarshalProto(notify)
			if err != nil {
				return err
			}
			u, err := notify.decode()
			if err != nil {
				log.Println(err)
				return err
			}
			event(u)
		}
		return nil
	}
	rsp := new(SubResponse)
	err := conn.Rpc(byte(Command_CMD_SUB), &Sub{Topics: []string{"private"}}, rsp)
	if err != nil {
		log.Println(err)
		t.TredeEvent(event)
		return
	}
	log.Println(rsp)
}
