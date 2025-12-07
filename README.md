# LongPort Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/gowsp/longport.svg)](https://pkg.go.dev/github.com/gowsp/longport)
[![Go Report Card](https://goreportcard.com/badge/github.com/gowsp/longport)](https://goreportcard.com/report/github.com/gowsp/longport)

LongPort Go SDK 是一个为 Go 语言开发者提供的 LongPort OpenAPI 客户端库，允许开发者轻松访问 LongPort 的各种金融功能。

## 功能特性

- 简单易用的 LongPort OpenAPI 认证
- RESTful API 客户端用于常规操作
- WebSocket 连接用于实时行情和交易更新
- 结构化的数据模型用于所有 API 响应
- 全面的错误处理机制
- 支持多种订单类型和交易功能

## 环境要求

- Go 1.25 或更高版本

## 安装

要安装该包，请运行：

```bash
$ go get github.com/gowsp/longport
```

## 快速开始

### 导入包

```go
import "github.com/gowsp/longport"
```

### 创建客户端

要与 API 进行交互，您需要使用您的凭据创建一个 LongPort 客户端：

```go
client := &longport.Longport{
    Host:        "longportapp.com", // 或 longportapp.cn
    AppKey:      "your-app-key",
    AppSecret:   "your-app-secret",
    AccessToken: "your-access-token",
}
```

### REST API 使用

#### 获取账户资产

```go
// 获取现金信息
cash, err := client.GetCash()
if err != nil {
    // 处理错误
}

// 获取指定币种的现金信息
usdCash, err := client.GetCash(longport.USD)
if err != nil {
    // 处理错误
}

// 获取股票持仓
stocks, err := client.GetStock()
if err != nil {
    // 处理错误
}

// 获取特定股票持仓
specificStocks, err := client.GetStock("AAPL.US", "GOOG.US")
if err != nil {
    // 处理错误
}
```

#### 订单操作

```go
// 提交订单
order := longport.SubmitOrder{
    BaseOrder: &longport.BaseOrder{
        Symbol:    "AAPL.US",
        OrderType: longport.LO,  // 限价单
        Side:      longport.Buy,
    },
    TimeInForce:       longport.Day,
    SubmittedQuantity: decimal.NewFromInt(10),
    SubmittedPrice:    "150.00",
}

rsp, err := client.SubmitOrder(order)
if err != nil {
    // 处理错误
}

// 修改订单
modifyOrder := longport.ModifyOrder{
    OrderID:  rsp.OrderID,
    Quantity: decimal.NewFromInt(20),
    Price:    "155.00",
}
err = client.ModifyOrder(modifyOrder)
if err != nil {
    // 处理错误
}

// 撤销订单
err = client.CancelOrder(rsp.OrderID)
if err != nil {
    // 处理错误
}

// 获取今日订单
orders, err := client.ListTodayOrder(longport.OrderQuery{
    Symbol: "AAPL.US",
    Side:   longport.Buy,
})
if err != nil {
    // 处理错误
}

// 获取历史订单
historyOrders, err := client.ListHistoryOrder(longport.HistoryQuery{
    OrderQuery: &longport.OrderQuery{
        Symbol: "AAPL.US",
    },
    Start: time.Now().AddDate(0, 0, -7).Unix(), // 7天前
    End:   time.Now().Unix(),
})
if err != nil {
    // 处理错误
}

// 预估最大购买数量
buyLimit, err := client.MaxOrderNum(longport.BuyLimitReq{
    Symbol:    "AAPL.US",
    OrderType: longport.LO,
    Side:      longport.Buy,
})
if err != nil {
    // 处理错误
}
```

### WebSocket 连接

#### 行情连接

```go
// 连接行情 WebSocket
quoteConn := client.ConnQuote()

// 查询标的证券信息
info, err := quoteConn.QuerySymbolStaticInfo("AAPL.US")
if err != nil {
    // 处理错误
}

// 查询多个标的证券信息
infos, err := quoteConn.QuerySymbolStaticInfo("AAPL.US", "GOOG.US")
if err != nil {
    // 处理错误
}

// 查询实时行情
quotes, err := quoteConn.QuerySymbolQuote("AAPL.US")
if err != nil {
    // 处理错误
}

// 订阅实时价格推送
quoteConn.OnPushQuote(func(quote *quotev1.PushQuote) {
    fmt.Printf("收到价格推送: %+v\n", quote)
})

// 订阅实时盘口推送
quoteConn.OnPushDepth(func(depth *quotev1.PushDepth) {
    fmt.Printf("收到盘口推送: %+v\n", depth)
})

// 订阅实时经纪队列推送
quoteConn.OnPushBrokers(func(brokers *quotev1.PushBrokers) {
    fmt.Printf("收到经纪队列推送: %+v\n", brokers)
})

// 订阅实时成交明细推送
quoteConn.OnPushTrade(func(trade *quotev1.PushTrade) {
    fmt.Printf("收到成交明细推送: %+v\n", trade)
})

// 订阅行情数据
subscribeResp, err := quoteConn.Subscribe(&quotev1.SubscribeRequest{
    Symbol: []string{"AAPL.US"},
    SubType: []quotev1.SubType{
        quotev1.SubType_QUOTE, 
        quotev1.SubType_DEPTH,
        quotev1.SubType_BROKER,
        quotev1.SubType_TRADE,
    },
})
if err != nil {
    // 处理错误
}

// 获取已订阅标的
subscriptionList, err := quoteConn.ListSubscription()
if err != nil {
    // 处理错误
}

// 取消订阅
err = quoteConn.Unsubscribe(&quotev1.UnsubscribeRequest{
    Symbol: []string{"AAPL.US"},
    SubType: []quotev1.SubType{
        quotev1.SubType_QUOTE, 
        quotev1.SubType_DEPTH,
    },
})
if err != nil {
    // 处理错误
}
```

#### 交易连接

```go
// 连接交易 WebSocket
tradeConn := client.ConnTrade()

// 订阅订单事件
err := tradeConn.Subscribe(func(event *longport.OrderEvent) {
    // 处理订单事件
    fmt.Printf("收到订单事件: %+v\n", event)
})
if err != nil {
    // 处理错误
}

// 取消订阅
err = tradeConn.Unsubscribe()
if err != nil {
    // 处理错误
}
```

## 支持的订单类型

SDK 支持以下订单类型：

- `LO`: 限价单
- `ELO`: 增强限价单
- `MO`: 市价单
- `AO`: 竞价市价单
- `ALO`: 竞价限价单
- `ODD`: 碎股单挂单
- `LIT`: 触价限价单
- `MIT`: 触价市价单
- `TSLPAMT`: 跟踪止损限价单 (跟踪金额)
- `TSLPPCT`: 跟踪止损限价单 (跟踪涨跌幅)
- `TSMAMT`: 跟踪止损市价单 (跟踪金额)
- `TSMPCT`: 跟踪止损市价单 (跟踪涨跌幅)
- `SLO`: 特殊限价单

## 支持的时间类型

- `Day`: 当日有效
- `GTC`: 撤单前有效
- `GTD`: 到期前有效

## API 文档

详细的 API 文档请参考 [官方 LongPort OpenAPI 文档](https://open.longportapp.com/zh-CN/docs)。

## 贡献

欢迎贡献！请随时提交 Pull Request。