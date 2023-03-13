package gone

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

// DEFAULT_CONFIG_PATH 默认配置文件地址
const DEFAULT_CONFIG_PATH string = "application.json"

type (
	GONE_CONFIG_CONTENT []byte // 配置文件内容，可通过 embed 嵌入
	GONE_TLS_CERT       string // https证书文件路径
	GONE_TLS_KEY        string // https密钥文件路径
)

const (
	gone_config int = iota
	gone_key_file
	gone_cert_file
)

// WebEngine http引擎，将作为http.Handler
type WebEngine struct {
	options map[int]any
	config  *AppConfig
	router  map[string]IHandler
}

// New 构造WebEngine
func New() *WebEngine {
	return &WebEngine{
		options: make(map[int]any),
		config:  newAppConfig(),
		router:  make(map[string]IHandler),
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

// Run 运行项目
func (engine *WebEngine) Run(opts ...any) {
	banner()
	engine.parseOptions(opts...)
	engine.parseConfig()
	engine.setLoggers()
	engine.Work()
}

// Work 开始工作
func (engine *WebEngine) Work() {
	basePath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if os.Getenv("GONE_ROUTINE") != "" || engine.config.Production {
		if err := engine.start(); err != nil {
			panic(err)
		}
	}
	watcher, err := NewWatcher(basePath, engine.config.Watcher.Includes, engine.config.Watcher.Excludes)
	if err != nil {
		LogWarn("error occurs when new watcher, your hot rebuild may not work")
	}
	engine.mainLoop(basePath, watcher)
}

// buildLoop 循环热更新
func (engine *WebEngine) mainLoop(basePath string, watcher *Watcher) {
	watchChan := make(chan int, 1)
	watcher.Start(watchChan)
	var cmd *exec.Cmd
	loadDone := make(chan int, 1)
	for {
		go Loading(loadDone, "Rebuilding...")
		if err := gbuild(basePath, engine.config.MainFile); err != nil {
			loadDone <- 1
			logWatch("error occurs when build project, ignore")
		} else {
			loadDone <- 1
			ClearTerm()
			if cmd != nil {
				cmd.Process.Kill()
			}
			cmd, err = grun(basePath)
			if err != nil {
				logWatch("error occurs when run project, ignore")
			}
		}
		<-watchChan
	}
}

// Start 启动 http 服务，如果条件允许，将启动 https 服务
func (engine *WebEngine) start() error {
	port := fmt.Sprintf(":%d", engine.config.Port)
	LogInfof("your serve is running on %s port\n", port)
	if engine.options[gone_cert_file] != nil && engine.options[gone_key_file] != nil {
		return http.ListenAndServeTLS(port, engine.options[gone_cert_file].(string), engine.options[gone_key_file].(string), engine)
	}
	return http.ListenAndServe(port, engine)
}

// parseOptions 读取项目变量
func (engine *WebEngine) parseOptions(opts ...any) {
	for _, opt := range opts {
		switch i := opt.(type) {
		// 配置内容
		case GONE_CONFIG_CONTENT:
			engine.options[gone_config] = []byte(i)
		// https证书
		case GONE_TLS_CERT:
			engine.options[gone_cert_file] = string(i)
		// https密钥
		case GONE_TLS_KEY:
			engine.options[gone_key_file] = string(i)
		}
	}
}

// parseConfig 加载配置文件
func (engine *WebEngine) parseConfig() {
	if content, ok := engine.options[gone_config]; ok {
		if err := engine.config.ParseContent(content.([]byte)); err != nil {
			LogWarn("error occurs when parse config from content, will use default config instead")
		} else {
			LogInfo("using config from content passed by user")
		}
	} else {
		if engine.config.IsValid(DEFAULT_CONFIG_PATH) {
			if err := engine.config.ParseFile(DEFAULT_CONFIG_PATH); err != nil {
				LogWarnf("error occurs when parse config from '%s', will use default config instead\n", DEFAULT_CONFIG_PATH)
			} else {
				LogInfof("using config from file '%s'\n", DEFAULT_CONFIG_PATH)
			}
		} else {
			LogInfo("config file not found, will use default config")
		}
	}
}

// setLoggers 设置输出
func (engine *WebEngine) setLoggers() {
	for _, each := range engine.config.Loggers {
		if gloggers[each.Name] != nil {
			gloggers[each.Name].SetStatus(each.Output)
		}
	}
}
