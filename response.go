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
	StatusCode int
	Data       string
}

// responseHtml 网页响应
type responseHtml struct {
	StatusCode int
	Data       []byte
}

// responseJson json响应
type responseJson struct {
	StatusCode int
	Data       any
}

// responseData 二进制响应
type responseData struct {
	StatusCode int
	Data       []byte
}

// String 构造字符串响应
func String(code int, format string, v ...any) *responseString {
	return &responseString{
		StatusCode: code,
		Data:       fmt.Sprintf(format, v...),
	}
}

// Html 构造网页响应
func Html(code int, data []byte) *responseHtml {
	return &responseHtml{
		StatusCode: code,
		Data:       data,
	}
}

// Json 构造 json 响应
func Json(code int, data any) *responseJson {
	return &responseJson{
		StatusCode: code,
		Data:       data,
	}
}

// Data 构造二进制响应
func Data(code int, data []byte) *responseData {
	return &responseData{
		StatusCode: code,
		Data:       data,
	}
}

// Invoke 执行字符串响应
func (response *responseString) Invoke(ctx *Context) {
	ctx.SetHeader("Content-Type", "text/plain; charset=utf-8")
	ctx.setStatusCode(response.StatusCode)
	ctx.Write([]byte(response.Data))
}

// Invoke 执行网页响应
func (response *responseHtml) Invoke(ctx *Context) {
	ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
	ctx.setStatusCode(response.StatusCode)
	ctx.Write(response.Data)
}

// Invoke 执行 json 响应
func (response *responseJson) Invoke(ctx *Context) {
	ctx.SetHeader("Content-Type", "application/json; charset=utf-8")
	ctx.setStatusCode(response.StatusCode)
	encoder := json.NewEncoder(ctx.Resp)
	if err := encoder.Encode(response.Data); err != nil {
		panic(err)
	}
}

// Invoke 执行二进制响应
func (response *responseData) Invoke(ctx *Context) {
	ctx.setStatusCode(response.StatusCode)
	ctx.Write(response.Data)
}
