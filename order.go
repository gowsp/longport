package longport

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// 订单接口
type OrderApi interface {
	// 获取历史订单
	ListHistoryOrder(query HistoryQuery) (*Orders, error)
	// 获取今日订单
	ListTodayOrder(query OrderQuery) (*Orders, error)
	//预估最大购买数量
	MaxOrderNum(req BuyLimitReq) (*BuyLimitRsp, error)
	// 提交订单
	SubmitOrder(req SubmitOrder) (*OrderRsp, error)
	// 修改订单
	ModifyOrder(modify ModifyOrder) error
	// 撤销订单
	CancelOrder(orderId string) error
}

type UnixTime time.Time

func (t *UnixTime) Day() string {
	return time.Time(*t).Format("20060102")
}
func (t *UnixTime) Add(d time.Duration) string {
	return time.Time(*t).Add(d).Format("2006-01-02")
}
func (t *UnixTime) UnmarshalJSON(data []byte) error {
	v := strings.Trim(string(data), `"`)
	val, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return err
	}
	*t = UnixTime(time.Unix(val, 0))
	return nil
}

// 基础订单信息
type BaseOrder struct {
	Symbol          string    `json:"symbol,omitempty"`           // required
	OrderType       OrderType `json:"order_type,omitempty"`       // required
	Side            OrderSide `json:"side,omitempty"`             // required
	TriggerPrice    string    `json:"trigger_price,omitempty"`    // LIT / MIT Order Required
	LimitOffset     string    `json:"limit_offset,omitempty"`     // TSLPAMT / TSLPPCT Order Required
	TrailingAmount  string    `json:"trailing_amount,omitempty"`  // TSLPAMT / TSMAMT Order Required
	TrailingPercent string    `json:"trailing_percent,omitempty"` // TSLPPCT / TSMAPCT Order Required
	Remark          string    `json:"remark,omitempty"`
}

// 订单提交信息
type SubmitOrder struct {
	*BaseOrder
	TimeInForce       TimeType        `json:"time_in_force,omitempty"`      // required
	ExpireDate        string          `json:"expire_date,omitempty"`        // required when time_in_force is GTD
	SubmittedQuantity decimal.Decimal `json:"submitted_quantity,omitempty"` // required
	SubmittedPrice    string          `json:"submitted_price,omitempty"`    // LO / ELO / ALO / ODD / LIT Order Required
}

// 订单信息
type Order struct {
	*CommonOrder
	LastDone    string   `json:"last_done,omitempty"`
	TimeInForce TimeType `json:"time_in_force,omitempty"`
	ExpireDate  string   `json:"expire_date,omitempty"`
	OutsideRth  string   `json:"outside_rth,omitempty"`
}

// 订单公共信息
type CommonOrder struct {
	*BaseOrder
	StockName string `json:"stock_name,omitempty"`

	Quantity         decimal.Decimal `json:"quantity,omitempty"`
	ExecutedQuantity decimal.Decimal `json:"executed_quantity,omitempty"`
	ExecutedPrice    decimal.Decimal `json:"executed_price,omitempty"`
	Price            string          `json:"price,omitempty"`

	OrderID  string `json:"order_id,omitempty"`
	Currency string `json:"currency,omitempty"`

	Status        OrderStatus   `json:"status,omitempty"`
	TriggerStatus TriggerStatus `json:"trigger_status,omitempty"`

	SubmittedAt UnixTime `json:"submitted_at,omitempty"`
	UpdatedAt   UnixTime `json:"updated_at,omitempty"`

	Msg       string `json:"msg,omitempty"`
	Tag       string `json:"tag,omitempty"`
	TriggerAt string `json:"trigger_at,omitempty"`
}

type OrderRsp struct {
	OrderID string `json:"order_id"`
}

func (l *Longport) SubmitOrder(req SubmitOrder) (*OrderRsp, error) {
	rsp := new(OrderRsp)
	return rsp, l.Post("/v1/trade/order", req, rsp)
}

type ModifyOrder struct {
	OrderID         string          `json:"order_id,omitempty"`
	Quantity        decimal.Decimal `json:"quantity,omitempty"`
	Price           string          `json:"price,omitempty"`
	TriggerPrice    string          `json:"trigger_price,omitempty"`
	LimitOffset     decimal.Decimal `json:"limit_offset,omitempty"`
	TrailingAmount  string          `json:"trailing_amount,omitempty"`
	TrailingPercent string          `json:"trailing_percent,omitempty"`
	Remark          string          `json:"remark,omitempty"`
}

func (l *Longport) ModifyOrder(modify ModifyOrder) error {
	rsp := new(struct{})
	return l.Put("/v1/trade/order", modify, rsp)
}
func (l *Longport) CancelOrder(orderId string) error {
	req := struct {
		OrderId string `json:"order_id,omitempty"`
	}{OrderId: orderId}
	rsp := new(struct{})
	return l.Delete("/v1/trade/order", req, rsp)
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

type Orders struct {
	Orders  []Order `json:"orders"`
	HasMore bool    `json:"has_more"`
}

func (l *Longport) ListTodayOrder(query OrderQuery) (*Orders, error) {
	rsp := new(Orders)
	if err := l.Get("/v1/trade/order/today", rsp, query.param()); err != nil {
		return nil, err
	}
	return rsp, nil
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

func (l *Longport) ListHistoryOrder(query HistoryQuery) (*Orders, error) {
	rsp := new(Orders)
	if err := l.Get("/v1/trade/order/history", rsp, query.param()); err != nil {
		return nil, err
	}
	return rsp, nil
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
func (l *Longport) MaxOrderNum(req BuyLimitReq) (*BuyLimitRsp, error) {
	rsp := new(BuyLimitRsp)
	return rsp, l.Get("/v1/trade/estimate/buy_limit", rsp, req.param())
}
