package trade

import (
	"github.com/gowsp/longport/control"
	"github.com/shopspring/decimal"
)

type listRsp[T any] struct {
	List []T `json:"list"`
}

func NewAssets(api *control.Invoker) Assets {
	return &assets{api: api}
}

type Assets interface {
	// 现金
	Cash() (*Cash, error)
	// 持仓
	Stock() (*Stock, error)
}

type assets struct {
	api *control.Invoker
}

type Cash struct {
	TotalCash              decimal.Decimal `json:"total_cash"`
	MaxFinanceAmount       decimal.Decimal `json:"max_finance_amount"`
	RemainingFinanceAmount decimal.Decimal `json:"remaining_finance_amount"`
	RiskLevel              string          `json:"risk_level"`
	MarginCall             string          `json:"margin_call"`
	Currency               string          `json:"currency"`
	NetAssets              decimal.Decimal `json:"net_assets"`
	InitMargin             decimal.Decimal `json:"init_margin"`
	MaintenanceMargin      decimal.Decimal `json:"maintenance_margin"`
	CashInfos              []struct {
		WithdrawCash  decimal.Decimal `json:"withdraw_cash"`
		AvailableCash decimal.Decimal `json:"available_cash"`
		FrozenCash    decimal.Decimal `json:"frozen_cash"`
		SettlingCash  decimal.Decimal `json:"settling_cash"`
		Currency      string          `json:"currency"`
	} `json:"cash_infos"`
}

func (a *assets) Cash() (*Cash, error) {
	rsp := new(listRsp[Cash])
	if err := a.api.Get("/v1/asset/account", rsp, nil); err != nil {
		return nil, err
	}
	return &rsp.List[0], nil
}

type Stock struct {
	AccountChannel string `json:"account_channel"`
	StockInfo      []struct {
		Symbol            string          `json:"symbol"`
		SymbolName        string          `json:"symbol_name"`
		Currency          string          `json:"currency"`
		Quantity          decimal.Decimal `json:"quantity"`
		Market            string          `json:"market"`
		AvailableQuantity decimal.Decimal `json:"available_quantity"`
		CostPrice         decimal.Decimal `json:"cost_price,omitempty"`
		InitQuantity      decimal.Decimal `json:"init_quantity"`
	} `json:"stock_info"`
}

func (a *assets) Stock() (*Stock, error) {
	rsp := new(listRsp[Stock])
	if err := a.api.Get("/v1/asset/stock", rsp, nil); err != nil {
		return nil, err
	}
	return &rsp.List[0], nil
}
