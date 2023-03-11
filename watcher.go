package gone

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// DEFAULT_INOTIFY_MASK 默认监听事件
//
// IN_MODIFY 文件被修改
//
// IN_CREATE 目录被创建
//
// IN_DELETE_SELF 目录被删除
// TODO: 文件移动事件也应该触发热更新
const DEFAULT_INOTIFY_MASK uint32 = syscall.IN_MODIFY | syscall.IN_CREATE | syscall.IN_DELETE_SELF

// Watcher 文件监听
//
// 由于syscall只支持 linux 系统，所以该监控只在 linux 下有效，windows下无法使用
//
// 跨平台的解决方案似乎应该使用 golang.org/x/sys
type Watcher struct {
	fd         int
	basePath   string
	wds        map[string]int
	includes   map[string]int
	excludes   map[string]int
	events     chan *WEvent
	watchChan  chan int
	handleChan chan int
	lock       sync.Mutex
}

// WEvent 监听事件
type WEvent struct {
	Type int
	Msg  string
}

// NewWather 创建Watcher
//
// 默认监听 basePath 及其下所有目录，忽略隐藏文件。可通过 includes 和 excludes 来强制添加和排除监听
func NewWatcher(basePath string, includes []string, excludes []string) (*Watcher, error) {
	fd, err := syscall.InotifyInit()
	if err != nil {
		return nil, err
	}
	watcher := &Watcher{
		fd:         fd,
		basePath:   basePath,
		wds:        make(map[string]int),
		includes:   make(map[string]int),
		excludes:   make(map[string]int),
		events:     make(chan *WEvent, 1),
		watchChan:  make(chan int, 1),
		handleChan: make(chan int, 1),
		lock:       sync.Mutex{},
	}
	if err := watcher.checkDirpath(watcher.basePath); err != nil {
		return nil, err
	}
	// 添加附加文件
	for i, v := range includes {
		watcher.includes[filepath.Join(basePath, v)] = i
	}
	// 添加排除文件
	for i, v := range excludes {
		watcher.excludes[filepath.Join(basePath, v)] = i
	}
	return watcher, nil
}

// Start 开始监听
//
// 该方法会开启两个 goroutine，一个用于监听事件，一个用于处理事件。注意该方法是非阻塞的。
func (watcher *Watcher) Start(sigChan chan int) {
	dirs := []string{watcher.basePath}
	watcher.walkDir(watcher.basePath, &dirs)
	// 将当前目录及子目录加入监听
	for _, each := range dirs {
		if err := watcher.AddWatch(each); err != nil {
			watcher.events <- &WEvent{Type: -1, Msg: err.Error()}
		}
	}
	// 开启事件处理
	go watcher.handleLoop(sigChan)
	// 开启事件监听
	go watcher.watchLoop()
}

// AddWatch 将目录添加进监听
func (watcher *Watcher) AddWatch(dirpath string) error {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()
	if watcher.fd == -1 {
		return fmt.Errorf("%s: watcher has been closed", dirpath)
	}
	// 如果目录已被监听，不再重复添加
	if _, ok := watcher.wds[dirpath]; ok {
		return nil
	}
	// 如果目录路径不正确或目标路径不是目录，返回错误
	if err := watcher.checkDirpath(dirpath); err != nil {
		return err
	}
	// 添加监听
	wd, err := syscall.InotifyAddWatch(watcher.fd, dirpath, DEFAULT_INOTIFY_MASK)
	if err != nil {
		return fmt.Errorf("%s: add watch error: %v", dirpath, err)
	}
	// 缓存watchdesc
	watcher.wds[dirpath] = wd
	return nil
}

// DeleteWatch 删除监听
func (watcher *Watcher) DeleteWatch(dirpath string) {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()
	// 当目录被删除时会自动移除该监听，所以只需要将其缓存删除
	delete(watcher.wds, dirpath)
}

// Close 关闭监听
func (watcher *Watcher) Close() {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()
	// 结束监听进程和处理进程
	watcher.watchChan <- 1
	watcher.handleChan <- 1
	// 关闭未关闭的监听
	for dirpath, wd := range watcher.wds {
		syscall.InotifyRmWatch(watcher.fd, uint32(wd))
		delete(watcher.wds, dirpath)
	}
	// 关闭通知
	syscall.Close(watcher.fd)
	watcher.fd = -1
}

// watchLoop 监听循环
func (watcher *Watcher) watchLoop() {
	var buf [syscall.SizeofInotifyEvent * 4096]byte
	for {
		select {
		// 监听到退出信号
		case <-watcher.watchChan:
			return
		default:
			n, err := syscall.Read(watcher.fd, buf[:])
			if err != nil {
				watcher.events <- &WEvent{Type: -1, Msg: fmt.Sprintf("syscall.Read error: %v", err)}
			}
			// 未监听到有效事件
			if n < syscall.SizeofInotifyEvent {
				continue
			}
			var offset uint32
			for offset <= uint32(n-syscall.SizeofInotifyEvent) {
				raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
				bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
				name := strings.TrimRight(string(bytes[0:raw.Len]), "\000")
				name_path := watcher.getFullPath(name, int(raw.Wd))
				// 无法获取文件完整路径，不在监听缓存内，忽略
				if name_path == "" {
					continue
				}
				// 监听到文件修改事件
				if raw.Mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
					watcher.events <- &WEvent{Type: syscall.IN_MODIFY, Msg: name_path}
				}
				// 监听到 文件/目录 创建事件
				if raw.Mask&syscall.IN_CREATE == syscall.IN_CREATE {
					if err := watcher.checkDirpath(name_path); err == nil {
						watcher.events <- &WEvent{Type: syscall.IN_CREATE, Msg: name_path}
					}
				}
				// 监听到被监听对象删除（目录删除）
				if raw.Mask&syscall.IN_DELETE_SELF == syscall.IN_DELETE_SELF {
					watcher.events <- &WEvent{Type: syscall.IN_DELETE_SELF, Msg: name_path}
				}
				offset += syscall.SizeofInotifyEvent + raw.Len
			}
		}
	}
}

// handleLoop 事件处理循环
func (watcher *Watcher) handleLoop(sigChan chan int) {
	debounce := Debounce(time.Millisecond * 100)
	for {
		select {
		case <-watcher.handleChan:
			return
		case event := <-watcher.events:
			// 处理文件修改
			if event.Type == syscall.IN_MODIFY {
				if watcher.checkValid(event.Msg) {
					debounce(func() {
						sigChan <- 1
					})
				}
			}
			// 处理目录创建
			if event.Type == syscall.IN_CREATE {
				fmt.Printf("%s: 文件被创建\n", event.Msg)
				if watcher.checkValid(event.Msg) {
					if err := watcher.AddWatch(event.Msg); err != nil {
						watcher.events <- &WEvent{Type: -1, Msg: fmt.Sprintf("add watch error: %v", err)}
					}
				}
			}
			// 处理目录删除
			if event.Type == syscall.IN_DELETE_SELF {
				watcher.DeleteWatch(event.Msg)
				fmt.Printf("%s: 文件被删除\n", event.Msg)
			}
			// 处理监听错误
			if event.Type == -1 {
				fmt.Println(event.Msg)
			}
		}
	}
}

// checkDirpath 检查目录路径
func (watcher *Watcher) checkDirpath(dirpath string) error {
	stat, err := os.Stat(dirpath)
	// 目录不存在
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("%s: no such directory", dirpath)
	}
	// 获取目录信息错误
	if err != nil {
		return fmt.Errorf("%s: %v", dirpath, err)
	}
	// 目标不是目录
	if !stat.IsDir() {
		return fmt.Errorf("%s: only directory can be added in watcher, not file", dirpath)
	}
	return nil
}

// walkDir 遍历目录及子目录
func (watcher *Watcher) walkDir(dirpath string, results *[]string) {
	dirs, _ := os.ReadDir(dirpath)
	for _, dir := range dirs {
		if dir.IsDir() {
			full_path := filepath.Join(dirpath, dir.Name())
			// 隐藏目录，判断是否被包含
			if strings.HasPrefix(dir.Name(), ".") {
				if _, ok := watcher.includes[full_path]; !ok {
					continue
				}
			}
			// 被排除目录
			if _, ok := watcher.excludes[full_path]; ok {
				continue
			}
			*results = append(*results, full_path)
			watcher.walkDir(full_path, results)
		}
	}
}

// getFullPath 根据监听获得完整路径
func (watcher *Watcher) getFullPath(name string, wd int) string {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()
	for key, value := range watcher.wds {
		if value == wd {
			return filepath.Join(key, name)
		}
	}
	return ""
}

// checkDir 检查目录是否应该添加进监控
func (watcher *Watcher) checkValid(name string) bool {
	if _, ok := watcher.excludes[name]; ok {
		return false
	}
	if _, ok := watcher.includes[name]; ok {
		return true
	}
	tmps := strings.Split(name, "/")
	return !strings.HasPrefix(tmps[len(tmps)-1], ".")
}
