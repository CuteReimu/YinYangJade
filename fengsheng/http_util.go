package fengsheng

import (
	"errors"
	"fmt"
	. "github.com/CuteReimu/onebot"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"time"
)

var restyClient = resty.New()

func init() {
	restyClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	restyClient.SetTimeout(20 * time.Second)
	restyClient.SetHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69",
		"connection":   "close",
	})
}

type errorWithMessage struct {
	error
	message MessageChain
}

func httpGet(endPoint string, queryParams map[string]string) *errorWithMessage {
	resp, err := restyClient.R().SetQueryParams(queryParams).Get(fengshengConfig.GetString("fengshengUrl") + endPoint)
	if err != nil {
		return &errorWithMessage{error: err}
	}
	if resp.StatusCode() != 200 {
		err := fmt.Errorf("请求错误，错误码：%d", resp.StatusCode())
		return &errorWithMessage{error: err, message: MessageChain{&Text{Text: err.Error()}}}
	}
	body := resp.String()
	if !gjson.Valid(body) {
		return &errorWithMessage{error: errors.New("json unmarshal failed")}
	}
	if returnError := gjson.Get(body, "error"); returnError.Exists() {
		s := returnError.String()
		return &errorWithMessage{error: errors.New(s), message: MessageChain{&Text{Text: s}}}
	}
	return nil
}

func httpGetBool(endPoint string, queryParams map[string]string) (bool, *errorWithMessage) {
	resp, err := restyClient.R().SetQueryParams(queryParams).Get(fengshengConfig.GetString("fengshengUrl") + endPoint)
	if err != nil {
		return false, &errorWithMessage{error: err}
	}
	if resp.StatusCode() != 200 {
		err := fmt.Errorf("请求错误，错误码：%d", resp.StatusCode())
		return false, &errorWithMessage{error: err, message: MessageChain{&Text{Text: err.Error()}}}
	}
	body := resp.String()
	if !gjson.Valid(body) {
		return false, &errorWithMessage{error: errors.New("json unmarshal failed")}
	}
	if returnError := gjson.Get(body, "error"); returnError.Exists() {
		s := returnError.String()
		return false, &errorWithMessage{error: errors.New(s), message: MessageChain{&Text{Text: s}}}
	}
	return gjson.Get(body, "result").Bool(), nil
}

func httpGetString(endPoint string, queryParams map[string]string) (string, *errorWithMessage) {
	resp, err := restyClient.R().SetQueryParams(queryParams).Get(fengshengConfig.GetString("fengshengUrl") + endPoint)
	if err != nil {
		return "", &errorWithMessage{error: err}
	}
	if resp.StatusCode() != 200 {
		err := fmt.Errorf("请求错误，错误码：%d", resp.StatusCode())
		return "", &errorWithMessage{error: err, message: MessageChain{&Text{Text: err.Error()}}}
	}
	body := resp.String()
	if !gjson.Valid(body) {
		return "", &errorWithMessage{error: errors.New("json unmarshal failed")}
	}
	if returnError := gjson.Get(body, "error"); returnError.Exists() {
		s := returnError.String()
		return "", &errorWithMessage{error: errors.New(s), message: MessageChain{&Text{Text: s}}}
	}
	return gjson.Get(body, "result").String(), nil
}
