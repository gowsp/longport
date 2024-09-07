package trade

import (
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type UnixTime time.Time

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

func (o *CommonOrder) IsBuy() bool {
	return o.Side == BUY_SIDE
}
func (o *CommonOrder) IsFilled() bool {
	return o.Status == FilledStatus || o.Status == PartialFilledStatus
}
