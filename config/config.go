package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

type ConnType string

const (
	QUOTE ConnType = "quote"
	TRADE ConnType = "trade"
)

type Config struct {
	Host        string `json:"url,omitempty"`
	AppKey      string `json:"app_key,omitempty"`
	AppSecret   string `json:"app_secret,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

func (c *Config) Conn(ct ConnType) string {
	return fmt.Sprintf("wss://openapi-%s.%s?version=1&codec=1&platform=9", c.Host, ct)
}
func (c *Config) Path(path string) string {
	return "https://openapi." + c.Host + path
}
func (c *Config) Hmac(plantext string) []byte {
	h := hmac.New(sha256.New, []byte(c.AppSecret))
	h.Write([]byte(plantext))
	return h.Sum(nil)
}
