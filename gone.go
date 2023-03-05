package gone

import "net/http"

type Gone struct {
	router map[string]IHandler
}

// New Gone的构造函数
func New() *Gone {
	return &Gone{router: make(map[string]IHandler)}
}

// addRoute 添加路由
func (gone *Gone) addRoute(method, pattern string, handler IHandler) {
	gone.router[method+"-"+pattern] = handler
}

// Get 添加 GET 路由
func (gone *Gone) Get(pattern string, handler func(ctx *Context) IResponse) {
	gone.addRoute("GET", pattern, Handler(handler))
}

// Post 添加 POST 路由
func (gone *Gone) Post(pattern string, handler func(ctx *Context) IResponse) {
	gone.addRoute("POST", pattern, Handler(handler))
}

// Run 启动服务
func (gone *Gone) Run(addr string) error {
	return http.ListenAndServe(addr, gone)
}

func (gone *Gone) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
