package longport

import (
	quotev1 "github.com/longportapp/openapi-protobufs/gen/go/quote"
	"google.golang.org/protobuf/proto"
)

// 行情长连接
type QuoteConn interface {
	// 可自定义执行 quotev1.Command 命令
	Rpc(cmd quotev1.Command, req proto.Message, rsp proto.Message) error
	// 查询交易时间段
	QueryMarketTradePeriod() (*quotev1.MarketTradePeriodResponse, error)
	// 查询标的基础信息
	QuerySymbolStaticInfo(symbols ...string) (*quotev1.SecurityStaticInfoResponse, error)
}
type quoteConn struct {
	*websocket
}

func (w *quoteConn) Rpc(cmd quotev1.Command, req proto.Message, rsp proto.Message) error {
	return w.websocket.Rpc(byte(cmd), req, rsp)
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
