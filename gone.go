package gone

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
func (engine *GEngine) parseConfig() *AppConfig {
	config := newAppConfig()
	err := config.ParseFile(engine.configPath)
	if err == nil {
		setLogger(config)
		LogInfof("Using config from file '%s'", engine.configPath)
		return config
	}
	if err != nil && os.IsNotExist(err) {
		LogWarnf("Config file '%s' not found, using default config.", engine.configPath)
		return config
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
	banner()
	config := engine.parseConfig()
	// 生产环境或编译后的进程，运行项目
	if config.Production || os.Getenv("GONE_RUNTIME") != "" {
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
			rebuildChan := make(chan int, 1)
			buildChan := make(chan int, 1)
			if err := startWatch(dir, watcher, rebuildChan, buildChan); err != nil {
				LogWarn("Unable to start watcher. Your hot build may not work.")
			}
			// 开始循环重构
			var cmd *exec.Cmd
			if build(dir) {
				cmd = run(dir)
				for {
					clearTerminal()
					<-rebuildChan
					if build(dir) {
						cmd.Process.Kill()
						buildChan <- 1
						cmd = run(dir)
					}
				}
			}
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
func startWatch(basePath string, watcher *Watcher, rebuildChan chan int, buildChan chan int) error {
	dirs := []string{basePath}
	getDirs(basePath, &dirs)
	for _, dir := range dirs {
		if err := watcher.AddWatch(dir); err != nil {
			return err
		}
		logWatchf("%s add to watcher successfully.", dir)
	}
	go watcher.Watch()
	go watcherEventsHandler(watcher, rebuildChan, buildChan)
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
func watcherEventsHandler(watcher *Watcher, rebuildChan chan int, buildChan chan int) {
	f := debounce(time.Millisecond * 300)
	spinner := GetSpinner()
	for {
		event := <-watcher.events
		switch event.Type {
		// TODO: 重新编译输出
		case syscall.IN_MODIFY:
			f(func() {
				rebuildChan <- 1
				go func() {
					for {
						select {
						case <-buildChan:
							return
						default:
							fmt.Printf("\033[1;32m\r %s Rebuilding... \033[0m", spinner())
							time.Sleep(time.Millisecond * 100)
						}
					}
				}()
			})
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

// clearTerminal 清屏
func clearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// debounce 防抖
func debounce(after time.Duration) func(func()) {
	var timer *time.Timer
	return func(f func()) {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(after, f)
	}
}

// banner 输出Banner
func banner() {
	fmt.Println(" \033[1;32m   ______   \033[1;36m____     \033[1;33m_   __    \033[1;31m______\033[0m")
	fmt.Println(" \033[1;32m  / ____/  \033[1;36m/ __ \\   \033[1;33m/ | / /   \033[1;31m/ ____/\033[0m")
	fmt.Println(" \033[1;32m / / __   \033[1;36m/ / / /  \033[1;33m/  |/ /   \033[1;31m/ __/   \033[0m")
	fmt.Println(" \033[1;32m/ /_/ /  \033[1;36m/ /_/ /  \033[1;33m/ /|  /   \033[1;31m/ /___   \033[0m")
	fmt.Println(" \033[1;32m\\____/   \033[1;36m\\____/  \033[1;33m/_/ |_/   \033[1;31m/_____/   \033[0m")
	fmt.Println()
}
