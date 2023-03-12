package gone

import "net/http"

// WebEngine http引擎，将作为http.Handler
type WebEngine struct {
	router map[string]IHandler
}

// New 构造WebEngine
func New() *WebEngine {
	return &WebEngine{
		router: make(map[string]IHandler),
	}
}

// addRoute 添加路由
func (engine *WebEngine) addRoute(method, pattern string, handler IHandler) {
	engine.router[method+"-"+pattern] = handler
}

// Get 添加 GET 路由
func (engine *WebEngine) Get(pattern string, handler func(ctx *Context) IResponse) {
	engine.addRoute("GET", pattern, Handler(handler))
}

// Post 添加 POST 路由
func (engine *WebEngine) Post(pattern string, handler func(ctx *Context) IResponse) {
	engine.addRoute("POST", pattern, Handler(handler))
}

// ServeHTTP 实现http.Handler
func (engine *WebEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r)
	if handler, ok := engine.router[ctx.Method+"-"+ctx.Path]; ok {
		handler.Do(ctx).Invoke(ctx)
	} else {
		String(http.StatusNotFound, "404 Not Found: %s", ctx.Path)
	}
}
