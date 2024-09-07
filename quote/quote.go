package quote

import (
	"time"

	"github.com/gowsp/longport/config"
	"github.com/gowsp/longport/control"
	"github.com/gowsp/talib"
	"github.com/shopspring/decimal"
)

type Quote interface {
	// 是否交易日
	TradingDays(market string, begin, end time.Time) (*MarketTradeDayResponse, error)
	TradingSession() (*MarketTradePeriodResponse, error)
	// 订阅行情
	Subscribe(symbols []string) error
	// K线
	Series(symbol string, period Period) (*talib.TimeSeries, error)
	// 静态信息
	Static(symbols []string) (map[string]*StaticInfo, error)
}

func New(api *control.Invoker) Quote {
	conn := control.NewConn(api, config.QUOTE)
	q := &quote{conn: conn}
	conn.Event = q.event
	return q
}

type quote struct {
	conn *control.Conn
}

func (w *quote) Series(symbol string, period Period) (*talib.TimeSeries, error) {
	rsp, err := w.candlestickImpl(symbol, period)
	if err != nil {
		return nil, err
	}
	periodTmp := time.Hour * 24
	series := talib.NewSeries(periodTmp)
	for _, v := range rsp.Candlesticks {
		bar := talib.NewBar(periodTmp,
			time.Unix(v.Timestamp, 0),
			decimal.RequireFromString(v.Open),
			decimal.RequireFromString(v.High),
			decimal.RequireFromString(v.Low),
			decimal.RequireFromString(v.Close),
		)
		series.Add(bar)
	}
	return series, nil
}
func (w *quote) event(push *control.Push) error {
	switch push.Cmd {
	case byte(Command_PushQuoteData):
		q := new(PushQuote)
		push.UnmarshalProto(q)
	}
	return nil
}
func (w *quote) TradingDays(market string, begin, end time.Time) (*MarketTradeDayResponse, error) {
	req := MarketTradeDayRequest{
		Market: market,
		BegDay: begin.Format("20060102"),
		EndDay: end.Format("20060102"),
	}
	rsp := new(MarketTradeDayResponse)
	return rsp, w.conn.Rpc(byte(Command_QueryMarketTradeDay), &req, rsp)
}
func (w *quote) TradingSession() (*MarketTradePeriodResponse, error) {
	rsp := new(MarketTradePeriodResponse)
	return rsp, w.conn.Rpc(byte(Command_QueryMarketTradePeriod), nil, rsp)
}
func (w *quote) Subscribe(symbols []string) error {
	req := SubscribeRequest{
		Symbol:      symbols,
		SubType:     []SubType{SubType_QUOTE},
		IsFirstPush: true,
	}
	rsp := new(SubscriptionResponse)
	return w.conn.Rpc(byte(Command_Subscribe), &req, rsp)
}
func (w *quote) Static(symbols []string) (map[string]*StaticInfo, error) {
	req := MultiSecurityRequest{
		Symbol: symbols,
	}
	rsp := new(SecurityStaticInfoResponse)
	err := w.conn.Rpc(byte(Command_QuerySecurityStaticInfo), &req, rsp)
	if err != nil {
		return nil, err
	}
	res := make(map[string]*StaticInfo)
	for _, info := range rsp.SecuStaticInfo {
		res[info.Symbol] = info
	}
	return res, nil
}
func (w *quote) candlestickImpl(symbol string, period Period) (*SecurityCandlestickResponse, error) {
	req := SecurityCandlestickRequest{
		Symbol:     symbol,
		Period:     period,
		AdjustType: AdjustType_NO_ADJUST,
		Count:      500,
	}
	rsp := new(SecurityCandlestickResponse)
	return rsp, w.conn.Rpc(byte(Command_QueryCandlestick), &req, rsp)
}
func (w *quote) HistoryCandlesticks(symbol string, period Period) (*talib.TimeSeries, error) {
	rsp, err := w.historyCandlesticksImpl(symbol, period)
	if err != nil {
		return nil, err
	}
	periodTmp := time.Hour * 24
	series := talib.NewSeries(periodTmp)
	for _, v := range rsp.Candlesticks {
		bar := talib.NewBar(periodTmp,
			time.Unix(v.Timestamp, 0),
			decimal.RequireFromString(v.Open),
			decimal.RequireFromString(v.High),
			decimal.RequireFromString(v.Low),
			decimal.RequireFromString(v.Close),
		)
		series.Add(bar)
	}
	return series, nil
}

func (w *quote) historyCandlesticksImpl(symbol string, period Period) (*SecurityCandlestickResponse, error) {
	req := SecurityHistoryCandlestickRequest{
		Symbol:        symbol,
		Period:        period,
		AdjustType:    AdjustType_NO_ADJUST,
		QueryType:     HistoryCandlestickQueryType_QUERY_BY_OFFSET,
		OffsetRequest: &SecurityHistoryCandlestickRequest_OffsetQuery{Direction: 0, Count: 1000},
	}
	rsp := new(SecurityCandlestickResponse)
	return rsp, w.conn.Rpc(byte(Command_QueryHistoryCandlestick), &req, rsp)
}
