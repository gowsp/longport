package longport

import (
	"net/url"

	"github.com/shopspring/decimal"
)

// 资产接口
type AssetApi interface {
	// 获取账户资金, 参数: 币种(可选)
	GetCash(c ...Currency) (*Cash, error)
	// 获取股票持仓, 参数: 股票代码(可选)，使用 ticker.region 格式，例如：AAPL.US
	GetStock(symbols ...string) (*Stocks, error)
}

type Stocks struct {
	List []struct {
		AccountChannel string  `json:"account_channel"`
		StockInfo      []Stock `json:"stock_info"`
	} `json:"list"`
}

type Stock struct {
	Symbol            string          `json:"symbol"`
	SymbolName        string          `json:"symbol_name"`
	Currency          Currency        `json:"currency"`
	Quantity          decimal.Decimal `json:"quantity"`
	Market            Market          `json:"market"`
	AvailableQuantity decimal.Decimal `json:"available_quantity"`
	CostPrice         decimal.Decimal `json:"cost_price,omitempty"`
	InitQuantity      decimal.Decimal `json:"init_quantity"`
}

func (l *Longport) GetStock(symbols ...string) (*Stocks, error) {
	list := new(Stocks)
	params := url.Values{}
	if len(symbols) > 0 {
		for _, symbol := range symbols {
			params.Add("symbol", symbol)
		}
	}
	err := l.Get("/v1/asset/stock", list, params)
	return list, err
}

type Cash struct {
	List []struct {
		TotalCash              decimal.Decimal `json:"total_cash"`
		MaxFinanceAmount       decimal.Decimal `json:"max_finance_amount"`
		RemainingFinanceAmount decimal.Decimal `json:"remaining_finance_amount"`
		RiskLevel              string          `json:"risk_level"`
		MarginCall             string          `json:"margin_call"`
		Currency               string          `json:"currency"`
		NetAssets              decimal.Decimal `json:"net_assets"`
		InitMargin             decimal.Decimal `json:"init_margin"`
		MaintenanceMargin      decimal.Decimal `json:"maintenance_margin"`
		BuyPower               decimal.Decimal `json:"buy_power"`
		CashInfos              []struct {
			WithdrawCash  decimal.Decimal `json:"withdraw_cash"`
			AvailableCash decimal.Decimal `json:"available_cash"`
			FrozenCash    decimal.Decimal `json:"frozen_cash"`
			SettlingCash  decimal.Decimal `json:"settling_cash"`
			Currency      Currency        `json:"currency"`
		} `json:"cash_infos"`
		FrozenTransactionFees []struct {
			Currency             Currency        `json:"currency"`
			FrozenTransactionFee decimal.Decimal `json:"frozen_transaction_fee"`
		} `json:"frozen_transaction_fees"`
	} `json:"list"`
}

func (l *Longport) GetCash(c ...Currency) (*Cash, error) {
	cash := new(Cash)
	params := url.Values{}
	if len(c) > 0 {
		params.Add("currency", string(c[0]))
	}
	err := l.Get("/v1/asset/account", cash, params)
	return cash, err
}
