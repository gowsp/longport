package longport

import (
	"log"
	"testing"

	quotev1 "github.com/longportapp/openapi-protobufs/gen/go/quote"
)

func TestQueryMarketTradePeriod(t *testing.T) {
	l := init_longport(t)
	val, err := l.ConnQuote().QueryMarketTradePeriod()
	log.Println(val, err)
}
func TestQuerySymbolStaticInfo(t *testing.T) {
	l := init_longport(t)
	val, err := l.ConnQuote().QuerySymbolStaticInfo("SPY.US", "QQQ.US")
	log.Println(val, err)
}
func TestQuerySecurityQuote(t *testing.T) {
	l := init_longport(t)
	val, err := l.ConnQuote().QuerySymbolQuote("SPY.US", "QQQ.US")
	log.Println(val, err)
}

func TestSubscribe(t *testing.T) {
	l := init_longport(t)
	quote := l.ConnQuote()
	quote.OnPushQuote(func(push *quotev1.PushQuote) {
		log.Println(push)
	})
	rsp, err := quote.Subscribe(&quotev1.SubscribeRequest{
		Symbol:      []string{"SPY.US", "QQQ.US"},
		SubType:     []quotev1.SubType{quotev1.SubType_QUOTE, quotev1.SubType_TRADE, quotev1.SubType_BROKERS, quotev1.SubType_DEPTH},
		IsFirstPush: true,
	})
	log.Println(rsp, err)
	quote.Subscribe(&quotev1.SubscribeRequest{
		Symbol:      []string{"SPYM.US", "QQQM.US"},
		SubType:     []quotev1.SubType{quotev1.SubType_QUOTE, quotev1.SubType_TRADE, quotev1.SubType_BROKERS, quotev1.SubType_DEPTH},
		IsFirstPush: true,
	})
	rsp, err = quote.ListSubscription()
	log.Println(rsp, err)
}
