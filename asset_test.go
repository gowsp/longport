package longport

import (
	"log"
	"testing"
)

func TestGetStock(t *testing.T) {
	l := init_longport(t)
	val, err := l.GetStock("SPY.US")
	log.Println(val, err)
}
func TestGetCash(t *testing.T) {
	l := init_longport(t)
	val, err := l.GetCash()
	log.Println(val, err)
}
