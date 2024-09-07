package control

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gowsp/longport/config"
)

type baseRsp struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func NewApi(config *config.Config) *Invoker {
	return &Invoker{config: config}
}

type Invoker struct {
	config *config.Config
}

func (i *Invoker) url(path string) string {
	return i.config.Path(path)
}
func (i *Invoker) do(method, path string, body, rsp any) error {
	req, err := http.NewRequest(method, i.url(path), nil)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json; charset=utf-8")
	i.sign(req, body)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if os.Getenv("LONGPORT_DEBUG") == "1" {
		data, _ := httputil.DumpResponse(res, true)
		fmt.Println(string(data))
	}
	bsp := new(baseRsp)
	bsp.Data = rsp
	if err = json.NewDecoder(res.Body).Decode(bsp); err != nil {
		return err
	}
	if bsp.Code == 0 {
		return nil
	}
	return fmt.Errorf("code: %d, msg: %s", bsp.Code, bsp.Message)
}
func (i *Invoker) sign(req *http.Request, body any) {
	headers := req.Header

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	access_token := i.config.AccessToken
	app_key := i.config.AppKey

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
	signature := hex.EncodeToString(i.config.Hmac(sign_str))
	headers.Set("X-Api-Signature", "HMAC-SHA256 SignedHeaders=authorization;x-api-key;x-timestamp, Signature="+signature)
}
func (i *Invoker) Get(path string, rsp any, params url.Values) error {
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}
	return i.do(http.MethodGet, path, nil, rsp)
}
func (i *Invoker) Put(path string, body, rsp any) error {
	return i.do(http.MethodPut, path, body, rsp)
}
func (i *Invoker) Post(path string, body, rsp any) error {
	return i.do(http.MethodPost, path, body, rsp)
}
func (i *Invoker) Delete(path string, body, rsp any) error {
	return i.do(http.MethodDelete, path, body, rsp)
}
