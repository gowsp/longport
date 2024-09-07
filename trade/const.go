package trade

type OrderSide string

const (
	BUY_SIDE  OrderSide = "Buy"
	SELL_SIDE OrderSide = "Sell"
)

type Market string

const (
	US_Market Market = "US"
	HK_Market Market = "HK"
)

func (m Market) Symbol(id string) string {
	return id + "." + string(m)
}

type OrderType string

const (
	LO      OrderType = "LO"      //限价单
	ELO     OrderType = "ELO"     //增强限价单
	MO      OrderType = "MO"      //市价单
	AO      OrderType = "AO"      //竞价市价单
	ALO     OrderType = "ALO"     //竞价限价单
	ODD     OrderType = "ODD"     //碎股单挂单
	LIT     OrderType = "LIT"     //触价限价单
	MIT     OrderType = "MIT"     //触价市价单
	TSLPAMT OrderType = "TSLPAMT" //跟踪止损限价单 (跟踪金额)
	TSLPPCT OrderType = "TSLPPCT" //跟踪止损限价单 (跟踪涨跌幅)
	TSMAMT  OrderType = "TSMAMT"  //跟踪止损市价单 (跟踪金额)
	TSMPCT  OrderType = "TSMPCT"  //跟踪止损市价单 (跟踪涨跌幅)
	SLO     OrderType = "SLO"     //特殊限价单，不支持改单
)

type TriggerStatus string

const (
	NOT_USED TriggerStatus = "NOT_USED" // 未激活
	DEACTIVE TriggerStatus = "DEACTIVE" // 已失效
	ACTIVE   TriggerStatus = "ACTIVE"   // 已激活
	RELEASED TriggerStatus = "RELEASED" // 已触发
)

type OrderStatus string

var NotReporteds = []OrderStatus{NotReported, ReplacedNotReported, ProtectedNotReported, VarietiesNotReported}

const (
	NotReported          OrderStatus = "NotReported"          //待提交
	ReplacedNotReported  OrderStatus = "ReplacedNotReported"  //待提交 (改单成功)
	ProtectedNotReported OrderStatus = "ProtectedNotReported" //待提交 (保价订单)
	VarietiesNotReported OrderStatus = "VarietiesNotReported" //待提交 (条件单)
	FilledStatus         OrderStatus = "FilledStatus"         //已成交
	WaitToNew            OrderStatus = "WaitToNew"            //已提待报
	NewStatus            OrderStatus = "NewStatus"            //已委托
	WaitToReplace        OrderStatus = "WaitToReplace"        //修改待报
	PendingReplaceStatus OrderStatus = "PendingReplaceStatus" //待修改
	ReplacedStatus       OrderStatus = "ReplacedStatus"       //已修改
	PartialFilledStatus  OrderStatus = "PartialFilledStatus"  //部分成交
	WaitToCancel         OrderStatus = "WaitToCancel"         //撤销待报
	PendingCancelStatus  OrderStatus = "PendingCancelStatus"  //待撤回
	RejectedStatus       OrderStatus = "RejectedStatus"       //已拒绝
	CanceledStatus       OrderStatus = "CanceledStatus"       //已撤单
	ExpiredStatus        OrderStatus = "ExpiredStatus"        //已过期
	PartialWithdrawal    OrderStatus = "PartialWithdrawal"    //部分撤单
)

type TimeType string

const (
	Day TimeType = "Day" // 当日有效
	GTC TimeType = "GTC" // 撤单前有效
	GTD TimeType = "GTD" // 到期前有效
)
