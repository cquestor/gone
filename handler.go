package gone

// IHandler 处理函数接口
type IHandler interface {
	Do(ctx *Context) IResponse
}

// Handler 处理函数
type Handler func(ctx *Context) IResponse

// Do 实现 IHandler 接口，函数式接口
func (handler Handler) Do(ctx *Context) IResponse {
	return handler(ctx)
}
