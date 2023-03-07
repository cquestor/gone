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

func (engine *GEngine) parseConfig() *AppConfig {
	config := newAppConfig()
	err := config.ParseFile(engine.configPath)
	if err == nil {
		LogInfof("Using config from file '%s'", engine.configPath)
		return config
	}
	if err != nil && os.IsNotExist(err) {
		LogWarnf("Config file '%s' not found, using default config", engine.configPath)
		return config
	}
	panic(err)
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

// Run 启动服务
func (engine *GEngine) Run() {
	config := engine.parseConfig()
	fmt.Println(config)
}

// ServeHTTP 实现 http.Handler 接口
func (engine *GEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
