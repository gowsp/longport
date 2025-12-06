package longport

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Method interface {
	Get(path string, rsp any, params url.Values) error
	Put(path string, body, rsp any) error
	Post(path string, body, rsp any) error
	Delete(path string, body, rsp any) error
}
type Api interface {
	Method
	OrderApi
	AssetApi
	// 连接行情 websocket
	ConnQuote() QuoteConn
	// 连接交易 websocket
	ConnTrade() TradeConn
}

// Create a custom HTTP client with timeout
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

type httpRsp struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func (r *httpRsp) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", r.Code, r.Message)
}

type Longport struct {
	Host        string `json:"host,omitempty"`
	AppKey      string `json:"app_key,omitempty"`
	AppSecret   string `json:"app_secret,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

func (l *Longport) ConnQuote() QuoteConn {
	url := fmt.Sprintf("wss://openapi-quote.%s?version=1&codec=1&platform=9", l.Host)
	return &quoteConn{websocket: &websocket{api: l, url: url}}
}
func (l *Longport) ConnTrade() TradeConn {
	url := fmt.Sprintf("wss://openapi-trade.%s?version=1&codec=1&platform=9", l.Host)
	return &tradeConn{websocket: &websocket{api: l, url: url}}
}

func (l *Longport) do(method, path string, body, rsp any) error {
	url := "https://openapi." + l.Host + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json; charset=utf-8")
	l.sign(req, body)

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	l.debug(req, res)
	result := new(httpRsp)
	result.Data = rsp
	if err = json.NewDecoder(res.Body).Decode(result); err != nil {
		return err
	}
	if result.Code == 0 {
		return nil
	}
	return result
}

func (l *Longport) debug(req *http.Request, res *http.Response) {
	if os.Getenv("DEBUG_LONGPORT") != "1" {
		return
	}
	log.Println("request:")
	data, _ := httputil.DumpRequestOut(req, true)
	log.Println(string(data))
	log.Println("response:")
	data, _ = httputil.DumpResponse(res, true)
	log.Println(string(data))
}
func (l *Longport) hmac(plantext string) []byte {
	h := hmac.New(sha256.New, []byte(l.AppSecret))
	h.Write([]byte(plantext))
	return h.Sum(nil)
}
func (l *Longport) sign(req *http.Request, body any) {
	headers := req.Header

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	access_token := l.AccessToken
	app_key := l.AppKey

	headers.Set("X-Api-Key", app_key)
	headers.Set("Authorization", access_token)
	headers.Set("X-Timestamp", ts)

	mtd := req.Method
	params := req.URL.Query().Encode()
	uri := req.URL.Path
	canonical_request := mtd + "|" + uri + "|" + params + "|authorization:" + access_token + "\nx-api-key:" + app_key + "\nx-timestamp:" + ts + "\n|authorization;x-api-key;x-timestamp|"
	if body != nil {
		data, _ := json.Marshal(body)
		req.Body = io.NopCloser(bytes.NewReader(data))
		d := sha1.Sum(data)
		canonical_request = canonical_request + hex.EncodeToString(d[:])
	}
	d := sha1.Sum([]byte(canonical_request))
	sign_str := "HMAC-SHA256|" + hex.EncodeToString(d[:])
	signature := hex.EncodeToString(l.hmac(sign_str))
	headers.Set("X-Api-Signature", "HMAC-SHA256 SignedHeaders=authorization;x-api-key;x-timestamp, Signature="+signature)
}
func (l *Longport) Get(path string, rsp any, params url.Values) error {
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}
	return l.do(http.MethodGet, path, nil, rsp)
}
func (l *Longport) Put(path string, body, rsp any) error {
	return l.do(http.MethodPut, path, body, rsp)
}
func (l *Longport) Post(path string, body, rsp any) error {
	return l.do(http.MethodPost, path, body, rsp)
}
func (l *Longport) Delete(path string, body, rsp any) error {
	return l.do(http.MethodDelete, path, body, rsp)
}
