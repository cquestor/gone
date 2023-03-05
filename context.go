package gone

import "net/http"

// Context 上下文
type Context struct {
	Req    *http.Request
	Writer http.ResponseWriter
	Method string
	Path   string
}

// newContext Context的构造函数
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Req:    r,
		Writer: w,
		Method: r.Method,
		Path:   r.URL.Path,
	}
}

// Write 写入响应数据
func (ctx *Context) Write(data []byte) {
	ctx.Writer.Write(data)
}

// SetStatusCode 写入响应状态码
func (ctx *Context) SetStatusCode(code int) {
	ctx.Writer.WriteHeader(code)
}

// SetHeader 设置响应头
func (ctx *Context) SetHeader(key, value string) {
	ctx.Writer.Header().Set(key, value)
}

// Query 获取请求参数
func (ctx *Context) Query(key string) string {
	return ctx.Req.URL.Query().Get(key)
}

// PostFrom 获取表单参数
func (ctx *Context) PostForm(key string) string {
	return ctx.Req.PostFormValue(key)
}

// Header 获取请求头
func (ctx *Context) Header(key string) string {
	return ctx.Req.Header.Get(key)
}
