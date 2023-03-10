package gone

import (
	"fmt"
	"os"
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
	fd   int
	wds  map[string]int
	lock sync.Mutex
}

// NewWather 创建Watcher
func NewWatcher() (*Watcher, error) {
	fd, err := syscall.InotifyInit()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		fd:   fd,
		wds:  make(map[string]int),
		lock: sync.Mutex{},
	}, nil
}

// Start 开始监听
func (watcher *Watcher) Start(basePath string) {

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
