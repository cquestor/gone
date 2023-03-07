package gone

import (
	"encoding/json"
	"fmt"
)

// IResponse 响应接口
type IResponse interface {
	Invoke(ctx *Context)
}

// responseString 字符串响应
type responseString struct {
	Code int
	Data string
}

// responseHtml 网页响应
type responseHtml struct {
	Code int
	Data []byte
}

// responseJson json响应
type responseJson struct {
	Code int
	Data any
}

// responseData 二进制响应
type responseData struct {
	Code int
	Data []byte
}

// String responseString构造函数
func String(code int, format string, values ...any) *responseString {
	return &responseString{
		Code: code,
		Data: fmt.Sprintf(format, values...),
	}
}

// Html responseHtml构造函数
func Html(code int, value []byte) *responseHtml {
	return &responseHtml{
		Code: code,
		Data: value,
	}
}

// Json responseJson构造函数
func Json(code int, value any) *responseJson {
	return &responseJson{
		Code: code,
		Data: value,
	}
}

// Data responseData构造函数
func Data(code int, value []byte) *responseData {
	return &responseData{
		Code: code,
		Data: value,
	}
}

// Invoke responseString
func (response *responseString) Invoke(ctx *Context) {
	ctx.SetHeader("Content-Type", "text/plain; charset=utf-8")
	ctx.SetStatusCode(response.Code)
	ctx.Write([]byte(response.Data))
}

// Invoke responseHtml
func (response *responseHtml) Invoke(ctx *Context) {
	ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
	ctx.SetStatusCode(response.Code)
	ctx.Write(response.Data)
}

// Invoke responseJson
func (response *responseJson) Invoke(ctx *Context) {
	ctx.SetHeader("Content-Type", "application/json; charset=utf-8")
	ctx.SetStatusCode(response.Code)
	encoder := json.NewEncoder(ctx.Writer)
	if err := encoder.Encode(response.Data); err != nil {
		panic(err)
	}
}

// Invoke responseData
func (response *responseData) Invoke(ctx *Context) {
	ctx.SetStatusCode(response.Code)
	ctx.Write(response.Data)
}
