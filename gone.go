package gone

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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
	setLogger(config)
	// 生产环境或编译后的进程，运行项目
	if config.Production || os.Getenv("GONE_RUNTIME") != "" {
		LogInfo(msg)
		if err := engine.start(fmt.Sprintf(":%d", config.Port)); err != nil {
			panic(err)
		}
	}
	watcher, err := NewWatcher()
	if err != nil {
		LogWarnf("Unable to new file watcher. Your hot build may not work.")
	} else {
		dir, err := os.Getwd()
		if err != nil {
			LogWarn("Unable to get the project path. Your hot build may not work.")
		} else {
			if err := startWatch(dir, watcher); err != nil {
				LogWarn("Unable to start watcher. Your hot build may not work.")
			}
			time.Sleep(10 * time.Minute)
		}
	}
}

// ServeHTTP 实现 http.Handler 接口
func (engine *GEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// setLogger 设置日志输出
func setLogger(config *AppConfig) {
	for _, logger := range config.Logger {
		switch strings.ToUpper(logger.Name) {
		case "INFO":
			loggerInfo.SetStatus(logger.Output)
		case "WARN":
			loggerWarn.SetStatus(logger.Output)
		case "ERROR":
			loggerError.SetStatus(logger.Output)
		case "WATCHER":
			loggerWatcher.SetStatus(logger.Output)
		}
	}
}

// startWatch 开始文件监测
func startWatch(basePath string, watcher *Watcher) error {
	dirs := []string{basePath}
	getDirs(basePath, &dirs)
	for _, dir := range dirs {
		if err := watcher.AddWatch(dir); err != nil {
			return err
		}
		logWatchf("%s add to watcher successfully.", dir)
	}
	go watcher.Watch()
	go watcherEventsHandler(watcher)
	build()
	return nil
}

// 获取目录
func getDirs(path string, dirs *[]string) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return
	}
	// TODO: watcher include and exclude
	for _, fi := range dir {
		if fi.IsDir() {
			if strings.HasPrefix(fi.Name(), ".") {
				continue
			}
			*dirs = append(*dirs, filepath.Join(path, fi.Name()))
			getDirs(filepath.Join(path, fi.Name()), dirs)
		}
	}
}

// watcherEventsHandler 处理文件事件
func watcherEventsHandler(watcher *Watcher) {
	for {
		event := <-watcher.events
		switch event.Type {
		case syscall.IN_MODIFY:
			logWatchf("%s Changed!", event.Name)
		case syscall.IN_CREATE:
			logWatchf("%s add to watcher successfully.", event.Name)
		case syscall.IN_DELETE_SELF:
			logWatchf("%s removed from watcher.", event.Name)
		case -1:
			LogWarnf("Watcher err: %s", event.Name)
			watcher.Close()
			LogWarn("Your hot build has closed!")
			return
		}
	}
}
