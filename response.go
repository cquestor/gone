package gone

type IResponse interface {
	Invoke(ctx *Context)
}
