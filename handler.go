package gone

// IHandler 处理函数接口
type IHandler interface {
	Invoke(ctx *Context) IResponse
}

// Handler 处理函数
type Handler func(ctx *Context) IResponse

// Invoke 实现 IHandler 接口
func (handler Handler) Invoke(ctx *Context) IResponse {
	return handler(ctx)
}
