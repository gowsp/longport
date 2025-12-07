package longport

import (
	"log"
	"sync"

	quotev1 "github.com/longportapp/openapi-protobufs/gen/go/quote"
	"google.golang.org/protobuf/proto"
)

// 行情长连接
type QuoteConn interface {
	// 可自定义执行 quotev1.Command 命令
	Rpc(cmd quotev1.Command, req proto.Message, rsp proto.Message) error

	// 注册实时价格回调函数
	OnPushQuote(func(*quotev1.PushQuote))
	// 注册实时盘口回调函数
	OnPushDepth(func(*quotev1.PushDepth))
	// 注册实时经纪队列回调函数
	OnPushBrokers(func(*quotev1.PushBrokers))
	// 注册实时成交明细回调函数
	OnPushTrade(func(*quotev1.PushTrade))
	// 订阅行情数据, 建议注册回调函数后调用
	Subscribe(req *quotev1.SubscribeRequest) (*quotev1.SubscriptionResponse, error)
	//取消订阅行情数据
	Unsubscribe(req *quotev1.UnsubscribeRequest) error
	// 获取已订阅标的行情
	ListSubscription() (*quotev1.SubscriptionResponse, error)

	// 查询交易时间段
	QueryMarketTradePeriod() (*quotev1.MarketTradePeriodResponse, error)
	// 查询标的基础信息
	QuerySymbolStaticInfo(symbols ...string) (*quotev1.SecurityStaticInfoResponse, error)
	// 获取标的实时行情
	QuerySymbolQuote(symbols ...string) (*quotev1.SecurityQuoteResponse, error)
}
type quoteConn struct {
	sync.Once
	*websocket

	handlers map[byte]func(proto.Message)
}

func (w *quoteConn) Rpc(cmd quotev1.Command, req proto.Message, rsp proto.Message) error {
	return w.websocket.Rpc(byte(cmd), req, rsp)
}
func (w *quoteConn) OnPushQuote(handler func(*quotev1.PushQuote)) {
	w.handlers[byte(quotev1.Command_PushQuoteData)] = func(m proto.Message) {
		handler(m.(*quotev1.PushQuote))
	}
}
func (w *quoteConn) OnPushDepth(handler func(*quotev1.PushDepth)) {
	w.handlers[byte(quotev1.Command_PushDepthData)] = func(m proto.Message) {
		handler(m.(*quotev1.PushDepth))
	}
}
func (w *quoteConn) OnPushBrokers(handler func(*quotev1.PushBrokers)) {
	w.handlers[byte(quotev1.Command_PushBrokersData)] = func(m proto.Message) {
		handler(m.(*quotev1.PushBrokers))
	}
}
func (w *quoteConn) OnPushTrade(handler func(*quotev1.PushTrade)) {
	w.handlers[byte(quotev1.Command_PushTradeData)] = func(m proto.Message) {
		handler(m.(*quotev1.PushTrade))
	}
}
func (w *quoteConn) Subscribe(req *quotev1.SubscribeRequest) (*quotev1.SubscriptionResponse, error) {
	w.Do(w.handle)
	rsp := new(quotev1.SubscriptionResponse)
	return rsp, w.Rpc(quotev1.Command_Subscribe, req, rsp)
}
func (w *quoteConn) ListSubscription() (*quotev1.SubscriptionResponse, error) {
	w.Do(w.handle)
	rsp := new(quotev1.SubscriptionResponse)
	return rsp, w.Rpc(quotev1.Command_Subscription, new(quotev1.SubscriptionRequest), rsp)
}
func (w *quoteConn) Unsubscribe(req *quotev1.UnsubscribeRequest) error {
	rsp := new(quotev1.UnsubscribeResponse)
	return w.Rpc(quotev1.Command_Unsubscribe, req, rsp)
}
func (w *quoteConn) handle() {
	w.websocket.Subscribe(func(event *event) error {
		conv := map[byte]func() proto.Message{
			byte(quotev1.Command_PushQuoteData):   func() proto.Message { return new(quotev1.PushQuote) },
			byte(quotev1.Command_PushDepthData):   func() proto.Message { return new(quotev1.PushDepth) },
			byte(quotev1.Command_PushBrokersData): func() proto.Message { return new(quotev1.PushBrokers) },
			byte(quotev1.Command_PushTradeData):   func() proto.Message { return new(quotev1.PushTrade) },
		}
		if newFunc, exists := conv[event.Cmd]; exists {
			data := newFunc()
			if err := event.UnmarshalProto(data); err != nil {
				return err
			}
			if handler := w.handlers[event.Cmd]; handler != nil {
				handler(data)
			}
		} else {
			log.Printf("unknown cmd: %d\n", event.Cmd)
		}
		return nil
	})
}
func (w *quoteConn) QueryMarketTradePeriod() (*quotev1.MarketTradePeriodResponse, error) {
	rsp := new(quotev1.MarketTradePeriodResponse)
	return rsp, w.Rpc(quotev1.Command_QueryMarketTradePeriod, nil, rsp)
}
func (w *quoteConn) QuerySymbolStaticInfo(symbols ...string) (*quotev1.SecurityStaticInfoResponse, error) {
	req := quotev1.MultiSecurityRequest{
		Symbol: symbols,
	}
	rsp := new(quotev1.SecurityStaticInfoResponse)
	err := w.Rpc(quotev1.Command_QuerySecurityStaticInfo, &req, rsp)
	return rsp, err
}
func (w *quoteConn) QuerySymbolQuote(symbols ...string) (*quotev1.SecurityQuoteResponse, error) {
	req := quotev1.MultiSecurityRequest{
		Symbol: symbols,
	}
	rsp := new(quotev1.SecurityQuoteResponse)
	err := w.Rpc(quotev1.Command_QuerySecurityQuote, &req, rsp)
	return rsp, err
}
