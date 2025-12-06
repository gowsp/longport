package longport

import (
	"log"
	"testing"
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
