package longport

import (
	"log"
	"testing"
)

func TestListHistoryOrder(t *testing.T) {
	api := init_longport(t)
	list, err := api.ListHistoryOrder(HistoryQuery{OrderQuery: &OrderQuery{Market: US, Symbol: "QQQ.US"}})
	log.Println(list, err)
}
func TestListTodayOrder(t *testing.T) {
	api := init_longport(t)
	list, err := api.ListTodayOrder(OrderQuery{Market: US, Symbol: "QQQ.US"})
	log.Println(list, err)
}
