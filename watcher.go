package gone

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
	fd       int
	basePath string
	wds      map[string]int
	includes map[string]int
	excludes map[string]int
	lock     sync.Mutex
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
		fd:       fd,
		basePath: basePath,
		wds:      make(map[string]int),
		includes: make(map[string]int),
		excludes: make(map[string]int),
		lock:     sync.Mutex{},
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
func (watcher *Watcher) Start() error {
	dirs := []string{watcher.basePath}
	watcher.walkDir(watcher.basePath, &dirs)
	// TODO: 开始事件监听
	fmt.Println(dirs)
	return nil
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
	// 关闭未关闭的监听
	for dirpath, wd := range watcher.wds {
		syscall.InotifyRmWatch(watcher.fd, uint32(wd))
		delete(watcher.wds, dirpath)
	}
	// 关闭通知
	syscall.Close(watcher.fd)
	watcher.fd = -1
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
