package longport

import (
	"github.com/gowsp/longport/config"
	"github.com/gowsp/longport/control"
	"github.com/gowsp/longport/quote"
	"github.com/gowsp/longport/trade"
)

type Api interface {
	quote.Quote
	trade.Assets
	trade.Trader
}

type app struct {
	trade.Assets
	trade.Trader
	quote.Quote
}

func New(config *config.Config) Api {
	api := control.NewApi(config)
	return &app{
		Quote:  quote.New(api),
		Assets: trade.NewAssets(api),
		Trader: trade.NewTrader(api),
	}
}
