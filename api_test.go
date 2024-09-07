package longport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gowsp/longport/config"
	"github.com/gowsp/longport/quote"
	"github.com/gowsp/longport/trade"
	"github.com/gowsp/talib"
	"github.com/shopspring/decimal"
)

func getApp() Api {
	return New(getConfig())
}

func getConfig() *config.Config {
	c := new(config.Config)
	// f, _ := os.Open("mock_test.json")
	f, _ := os.Open("real_test.json")
	json.NewDecoder(f).Decode(c)
	return c
}
func TestStock(t *testing.T) {
	a := getApp()
	list, _ := a.Stock()
	buff := new(bytes.Buffer)
	for _, stock := range list.StockInfo {
		symbol := stock.Symbol
		series, _ := a.Series(symbol, quote.Period_DAY)
		atr := series.Atr(100).Offset(0)
		atr2 := series.Atr(20).Offset(0)
		fmt.Println(atr2.Sub(atr).RoundCeil(3))
		close := series.ClosePriceIndicator().Offset(0)
		buff.WriteString(fmt.Sprintf("----- %s ----- %s -----\n", symbol, symbol))
		buff.WriteString(fmt.Sprintf("收盘: %s \n", close))
		highest := talib.Highest(series.HighPriceIndicator(), 20).Offset(0)
		buff.WriteString(fmt.Sprintf("最高: %s\t 波动: %s\t 波动率: %s%% \t\n", highest, atr.RoundCeil(3), atr.Div(close).Mul(talib.HUNDRED).RoundCeil(3)))
		buff.WriteString(fmt.Sprintf("差值: %v\t 止损: %s\t 上升: %s \n", highest.Sub(close), highest.Sub(atr.Mul(decimal.RequireFromString("3"))).RoundCeil(3), atr.Mul(decimal.RequireFromString("0.5")).RoundCeil(3)))
	}
	fmt.Println(buff.String())
}
func TestOrderEvent(t *testing.T) {
	app := getApp()
	app.TredeEvent(func(oe *trade.OrderEvent) {
		fmt.Println(oe)
	})
}

func TestSubmitOrder(t *testing.T) {
	base := trade.BaseOrder{
		Side:           trade.SELL_SIDE,
		Symbol:         "00883.HK",
		OrderType:      trade.TSLPAMT,
		TrailingAmount: decimal.NewFromInt(1).String(),
		LimitOffset:    decimal.Zero.String(),
	}
	o := trade.SubmitOrder{SubmittedQuantity: decimal.NewFromInt(1000), BaseOrder: &base, TimeInForce: trade.GTC}
	val, _ := json.Marshal(o)
	fmt.Println(string(string(val)))
	fmt.Println(getApp().Submit(o))
}

func TestCash(t *testing.T) {
	v, err := getApp().Cash()
	if err != nil {
		return
	}
	fmt.Println(v)
}

func TestTodayOrder(t *testing.T) {
	app := getApp()
	val, err := app.ListToday(trade.OrderQuery{Side: trade.BUY_SIDE})
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%v \n", val)
}
func TestHistory(t *testing.T) {
	app := getApp()
	q := &trade.OrderQuery{Status: []trade.OrderStatus{}}
	val, err := app.ListHistory(trade.HistoryQuery{OrderQuery: q})
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range val {
		if v.TimeInForce == trade.GTC {
			fmt.Println(v.OrderID, v.TriggerStatus, v.Status, v.ExecutedPrice, v.ExecutedQuantity)
		}
	}
}
func TestDelete(t *testing.T) {
	app := getApp()
	q := &trade.OrderQuery{Status: trade.NotReporteds, Symbol: "883.HK"}
	val, err := app.ListHistory(trade.HistoryQuery{OrderQuery: q})
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range val {
		if v.TimeInForce == trade.GTC {
			app.Cancel(v.OrderID)
		}
	}
}
func TestSubscribe(t *testing.T) {
	app := getApp()
	stocks, err := app.Stock()
	if err != nil {
		return
	}
	symbols := make([]string, 0)
	for _, s := range stocks.StockInfo {
		symbols = append(symbols, s.Symbol)
	}
	getApp().Subscribe(symbols)
	time.Sleep(time.Hour)
}
func TestModifyOrder(t *testing.T) {
	app := getApp()
	val, err := app.ListToday(trade.OrderQuery{})
	if err != nil {
		return
	}
	for _, v := range val {
		if v.TimeInForce != trade.GTC {
			continue
		}
		v.Quantity = decimal.NewFromInt(400)
		err := app.Modify(v.Modify())
		fmt.Println(err)
	}
}
func TestLimit(t *testing.T) {
	app := getApp()
	val, err := app.Limit(trade.BuyLimitReq{Symbol: "1810.HK", Side: trade.BUY_SIDE, OrderType: trade.MO})
	if err != nil {
		return
	}
	fmt.Println(val)
}
func TestStatic(t *testing.T) {
	app := getApp()
	val, err := app.Static([]string{"1810.HK", "1816.HK"})
	if err != nil {
		return
	}
	fmt.Println(val)
}
func TestWeb(t *testing.T) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Println(string(body))
		w.Header().Add("content-type", "application/json")
		fmt.Fprintf(w, `{"code": 0, "msg": "error"}`)
		w.WriteHeader(200)
	})
	http.ListenAndServe(":8080", nil)
}
