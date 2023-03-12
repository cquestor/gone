package gone

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

// Context 上下文
type Context struct {
	Req      *http.Request
	Resp     http.ResponseWriter
	Method   string
	Path     string
	lockCode sync.Once
}

// newContext 构造上下文
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Req:      r,
		Resp:     w,
		Method:   r.Method,
		Path:     r.URL.Path,
		lockCode: sync.Once{},
	}
}

// Header 获取请求头
func (ctx *Context) Header(key string) string {
	return ctx.Req.Header.Get(key)
}

// SetHeader 设置响应头
func (ctx *Context) SetHeader(key, value string) {
	ctx.Resp.Header().Set(key, value)
}

// Query 获取请求参数
func (ctx *Context) Query(key string) string {
	return ctx.Req.URL.Query().Get(key)
}

// PostFrom 获取表单参数
func (ctx *Context) PostForm(key string) string {
	return ctx.Req.PostFormValue(key)
}

// Body 获取请求体
func (ctx *Context) Body() []byte {
	b, _ := io.ReadAll(ctx.Req.Body)
	ctx.Req.Body = io.NopCloser(bytes.NewBuffer(b))
	return b
}

// Write 写入响应数据
func (ctx *Context) Write(data []byte) {
	ctx.Resp.Write(data)
}

// setStatusCode 设置响应状态码
func (ctx *Context) setStatusCode(code int) {
	ctx.lockCode.Do(func() {
		ctx.Resp.WriteHeader(code)
	})
}
