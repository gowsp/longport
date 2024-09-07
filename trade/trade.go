package trade

import (
	"net/url"
	"strconv"

	"github.com/gowsp/longport/control"
	"github.com/shopspring/decimal"
)

type Trader interface {
	// 提交订单
	Submit(req SubmitOrder) (*OrderRsp, error)
	// 修改订单
	Modify(modify ModifyOrder) error
	// 撤销订单
	Cancel(orderId string) error
	// 列出今日订单
	ListToday(query OrderQuery) ([]Order, error)
	// 列出历史订单
	ListHistory(query HistoryQuery) ([]Order, error)
	// 列出历史订单
	Limit(modify BuyLimitReq) (*BuyLimitRsp, error)
	// 订阅交易事件
	TredeEvent(event Event)
}

func NewTrader(api *control.Invoker) Trader {
	t := &trader{api: api}
	return t
}

type trader struct {
	api *control.Invoker
}
type Orders struct {
	Orders []Order `json:"orders"`
}

func (o *CommonOrder) Modify() ModifyOrder {
	req := ModifyOrder{OrderID: o.OrderID, Quantity: o.Quantity}
	switch o.OrderType {
	case LO, ELO, ALO, ODD, LIT:
		req.Price = o.Price
		if o.OrderType == LIT {
			req.TriggerPrice = o.TriggerPrice
		}
	case MIT:
		req.TriggerPrice = o.TriggerPrice
	case TSLPAMT, TSLPPCT:
		req.LimitOffset = decimal.RequireFromString(o.LimitOffset)
		if o.OrderType == TSLPAMT {
			req.TrailingAmount = o.TrailingAmount
		}
	case TSMAMT:
		req.TrailingAmount = o.TrailingAmount
	case TSMPCT:
		req.TrailingPercent = o.TrailingPercent
	}
	if o.Remark != "" {
		req.Remark = o.Remark
	}
	return req
}

type OrderQuery struct {
	Symbol string
	Side   OrderSide
	Market Market
	Status []OrderStatus
}

func (r *OrderQuery) param() url.Values {
	params := make(url.Values)
	if r.Symbol != "" {
		params.Set("symbol", r.Symbol)
	}
	if r.Side != "" {
		params.Set("side", string(r.Side))
	}
	if r.Market != "" {
		params.Set("market", string(r.Market))
	}
	if len(r.Status) > 0 {
		for _, status := range r.Status {
			if status == "" {
				continue
			}
			params.Add("status", string(status))
		}
	}
	return params
}

func (a *trader) ListToday(query OrderQuery) ([]Order, error) {
	rsp := new(Orders)
	if err := a.api.Get("/v1/trade/order/today", rsp, query.param()); err != nil {
		return nil, err
	}
	return rsp.Orders, nil
}

type HistoryQuery struct {
	*OrderQuery
	Start int64
	End   int64
}

func (r *HistoryQuery) param() url.Values {
	var params url.Values
	if r.OrderQuery == nil {
		params = make(url.Values)
	} else {
		params = r.OrderQuery.param()
	}
	if r.Start > 0 {
		params.Add("start_at", strconv.FormatInt(r.Start, 10))
	}
	if r.End > 0 {
		params.Add("end_at", strconv.FormatInt(r.End, 10))
	}
	return params
}

type HistoryOrder struct {
	Orders  []Order `json:"orders"`
	HasMore bool    `json:"has_more"`
}

func (a *trader) ListHistory(query HistoryQuery) ([]Order, error) {
	rsp := new(HistoryOrder)
	if err := a.api.Get("/v1/trade/order/history", rsp, query.param()); err != nil {
		return nil, err
	}
	return rsp.Orders, nil
}

func (a *trader) Modify(modify ModifyOrder) error {
	rsp := new(struct{})
	return a.api.Put("/v1/trade/order", modify, rsp)
}

func (a *trader) Cancel(orderId string) error {
	req := struct {
		OrderId string `json:"order_id,omitempty"`
	}{OrderId: orderId}
	rsp := new(struct{})
	return a.api.Delete("/v1/trade/order", req, rsp)
}

type BuyLimitRsp struct {
	CashMaxQty   string `json:"cash_max_qty"`
	MarginMaxQty string `json:"margin_max_qty"`
}

type BuyLimitReq struct {
	Symbol    string    `json:"symbol,omitempty"`
	OrderType OrderType `json:"order_type,omitempty"`
	Side      OrderSide `json:"side,omitempty"`
}

func (r *BuyLimitReq) param() url.Values {
	params := make(url.Values, 0)
	params.Set("symbol", r.Symbol)
	params.Set("order_type", string(r.OrderType))
	params.Set("side", string(r.Side))
	return params
}

func (a *trader) Limit(req BuyLimitReq) (*BuyLimitRsp, error) {
	rsp := new(BuyLimitRsp)
	return rsp, a.api.Get("/v1/trade/estimate/buy_limit", rsp, req.param())
}

type OrderRsp struct {
	OrderID string `json:"order_id"`
}

func (a *trader) Submit(req SubmitOrder) (*OrderRsp, error) {
	rsp := new(OrderRsp)
	return rsp, a.api.Post("/v1/trade/order", req, rsp)
}
