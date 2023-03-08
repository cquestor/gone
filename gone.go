package gone

import (
	"fmt"
	"net/http"
	"os"
)

// GEngine http.Handler
type GEngine struct {
	configPath string
	router     map[string]IHandler
}

// New GEngine构造函数
func New() *GEngine {
	return &GEngine{
		configPath: DEFAULT_CONFIG_PATH,
		router:     make(map[string]IHandler),
	}
}

// SetConfigPath 设置配置文件路径
func (engine *GEngine) SetConfigPath(path string) {
	engine.configPath = path
}

// parseConfig 加载配置文件
func (engine *GEngine) parseConfig() (*AppConfig, string) {
	config := newAppConfig()
	err := config.ParseFile(engine.configPath)
	if err == nil {
		return config, fmt.Sprintf("Using config from file '%s'", engine.configPath)
	}
	if err != nil && os.IsNotExist(err) {
		return config, fmt.Sprintf("Config file '%s' not found, using default config.", engine.configPath)
	}
	panic(fmt.Errorf("parse config from file '%s' error: %v", engine.configPath, err))
}

// addRoute 添加路由
func (engine *GEngine) addRoute(method, pattern string, handler IHandler) {
	engine.router[method+"-"+pattern] = handler
}

// Get 添加 Get 路由
func (engine *GEngine) Get(pattern string, handler func(ctx *Context) IResponse) {
	engine.addRoute("GET", pattern, Handler(handler))
}

// Post 添加 POST 路由
func (engine *GEngine) Post(pattern string, handler func(ctx *Context) IResponse) {
	engine.addRoute("POST", pattern, Handler(handler))
}

// start 启动服务
func (engine *GEngine) start(addr string) error {
	LogInfof("Server start on %v", addr)
	return http.ListenAndServe(addr, engine)
}

// Run 项目运行
func (engine *GEngine) Run() {
	config, msg := engine.parseConfig()
	// 生产环境或编译后的进程，运行项目
	if config.Production || os.Getenv("GONE_RUNTIME") != "" {
		LogInfo(msg)
		if err := engine.start(fmt.Sprintf(":%d", config.Port)); err != nil {
			panic(err)
		}
	}
}

// ServeHTTP 实现 http.Handler 接口
func (engine *GEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
