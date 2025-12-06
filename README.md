# LongPort Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/gowsp/longport.svg)](https://pkg.go.dev/github.com/gowsp/longport)
[![Go Report Card](https://goreportcard.com/badge/github.com/gowsp/longport)](https://goreportcard.com/report/github.com/gowsp/longport)

LongPort Go SDK 为 Go 语言应用提供了便捷的 LongPort API 访问方式。

## 功能特性

- 简单易用的 LongPort OpenAPI 认证
- RESTful API 客户端用于常规操作
- WebSocket 连接用于实时行情和交易更新
- 结构化的数据模型用于所有 API 响应
- 全面的错误处理机制

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

### 认证

要与 API 进行认证，您需要使用您的凭据创建一个 LongPort 客户端：

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

// 获取股票持仓
stocks, err := client.GetStock()
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
        OrderType: longport.LO,
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

// 撤销订单
err = client.CancelOrder("order-id")
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
```

## API 文档

详细的 API 文档请参考 [官方 LongPort OpenAPI 文档](https://open.longportapp.com/zh-CN/docs)。

## 贡献

欢迎贡献！请随时提交 Pull Request。