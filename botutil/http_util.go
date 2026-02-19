package botutil

import (
	"errors"
	"fmt"
	"time"

	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

// HTTPClient 是一个用于发送 HTTP 请求的客户端
type HTTPClient struct {
	urlPrefix   func() string
	RestyClient *resty.Client
}

// NewHTTPClient 创建一个 HTTPClient 实例
func NewHTTPClient(urlPrefix func() string) *HTTPClient {
	restyClient := resty.New()
	restyClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	restyClient.SetTimeout(20 * time.Second)
	restyClient.SetHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69",
		"connection":   "close",
	})
	return &HTTPClient{
		urlPrefix:   urlPrefix,
		RestyClient: restyClient,
	}
}

// ErrorWithMessage 包含错误信息和应当返回给用户的消息
type ErrorWithMessage struct {
	error
	Message MessageChain
}

func (c *HTTPClient) HTTPGet(endPoint string, queryParams map[string]string) *ErrorWithMessage {
	_, err := c.get(endPoint, queryParams)
	return err
}

func (c *HTTPClient) HTTPGetBool(endPoint string, queryParams map[string]string) (bool, *ErrorWithMessage) {
	body, err := c.get(endPoint, queryParams)
	if err != nil {
		return false, err
	}
	return gjson.Get(body, "result").Bool(), nil
}

func (c *HTTPClient) HTTPGetString(endPoint string, queryParams map[string]string) (string, *ErrorWithMessage) {
	body, err := c.get(endPoint, queryParams)
	if err != nil {
		return "", err
	}
	return gjson.Get(body, "result").String(), nil
}

func (c *HTTPClient) HTTPGetInt(endPoint string, queryParams map[string]string) (int64, *ErrorWithMessage) {
	body, err := c.get(endPoint, queryParams)
	if err != nil {
		return 0, err
	}
	return gjson.Get(body, "result").Int(), nil
}

func (c *HTTPClient) get(endPoint string, queryParams map[string]string) (string, *ErrorWithMessage) {
	resp, err := c.RestyClient.R().SetQueryParams(queryParams).Get(c.urlPrefix() + endPoint)
	if err != nil {
		return "", &ErrorWithMessage{error: err}
	}
	if resp.StatusCode() != 200 {
		err := fmt.Errorf("请求错误，错误码：%d", resp.StatusCode())
		return "", &ErrorWithMessage{error: err, Message: MessageChain{&Text{Text: err.Error()}}}
	}
	body := resp.String()
	if !gjson.Valid(body) {
		return "", &ErrorWithMessage{error: errors.New("json unmarshal failed")}
	}
	if returnError := gjson.Get(body, "error"); returnError.Exists() {
		s := returnError.String()
		return "", &ErrorWithMessage{error: errors.New(s), Message: MessageChain{&Text{Text: s}}}
	}
	return body, nil
}
