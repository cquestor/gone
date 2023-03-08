package gone

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// Watcher 文件监控
type Watcher struct {
	fd     int
	wds    map[string]int
	events chan *FEvent
	errs   chan *FEvent
	lock   sync.Mutex
}

// FEvent 文件事件
type FEvent struct {
	Name string
	Type int
}

// newWatcher Watcher构造函数
func NewWatcher() (*Watcher, error) {
	fd, err := syscall.InotifyInit()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		fd:     fd,
		wds:    make(map[string]int),
		events: make(chan *FEvent),
		lock:   sync.Mutex{},
	}, nil
}

// AddWatch 添加监听目录
func (watcher *Watcher) AddWatch(path string) error {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()
	if _, ok := watcher.wds[path]; ok {
		return nil
	}
	wd, err := syscall.InotifyAddWatch(watcher.fd, path, syscall.IN_MODIFY|syscall.IN_CREATE|syscall.IN_DELETE_SELF)
	if err != nil {
		return err
	}
	watcher.wds[path] = wd
	return nil
}

// RemoveWatch 删除监听
func (watcher *Watcher) RemoveWatch(path string) error {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()
	_, ok := watcher.wds[path]
	if !ok {
		return nil
	}
	delete(watcher.wds, path)
	return nil
}

// Watch 监听文件事件
func (watcher *Watcher) Watch() {
	var buf [syscall.SizeofInotifyEvent * 4096]byte
	for {
		n, err := syscall.Read(watcher.fd, buf[:])
		if err != nil {
			watcher.errs <- &FEvent{Name: "events read error", Type: -1}
		}
		if n < syscall.SizeofInotifyEvent {
			continue
		}
		var offset uint32
		for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			mask := uint32(raw.Mask)
			nameLen := uint32(raw.Len)
			bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
			name := strings.TrimRight(string(bytes[0:nameLen]), "\000")
			// FIXME: 文件路径可能为""
			name_path := watcher.GetPath(name, int(raw.Wd))
			// 文件修改
			if mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
				watcher.events <- &FEvent{Name: name_path, Type: syscall.IN_MODIFY}
			}
			// 文件创建
			if mask&syscall.IN_CREATE == syscall.IN_CREATE {
				info, err := os.Stat(name_path)
				if err != nil {
					watcher.errs <- &FEvent{Name: fmt.Sprintf("watch add (os.Stat) error: %v", err.Error()), Type: -1}
				}
				if info.IsDir() {
					err := watcher.AddWatch(name_path)
					if err != nil {
						watcher.errs <- &FEvent{Name: fmt.Sprintf("watch add (AddWatch) error: %v", err.Error()), Type: -1}
					}
					watcher.events <- &FEvent{Name: name_path, Type: syscall.IN_CREATE}
				}
			}
			// 目录删除
			if mask&syscall.IN_DELETE_SELF == syscall.IN_DELETE_SELF {
				if err := watcher.RemoveWatch(name_path); err != nil {
					watcher.errs <- &FEvent{Name: "watch remove error", Type: -1}
				}
				watcher.events <- &FEvent{Name: name_path, Type: syscall.IN_DELETE_SELF}
			}
			offset += syscall.SizeofInotifyEvent + nameLen
		}
	}
}

// GetPath 获取文件路径
func (watcher *Watcher) GetPath(name string, wd int) string {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()
	for key, value := range watcher.wds {
		if value == wd {
			return path.Join(key, name)
		}
	}
	return ""
}

// Close 关闭检测
func (watcher *Watcher) Close() {
	for _, wd := range watcher.wds {
		syscall.InotifyRmWatch(watcher.fd, uint32(wd))
	}
	syscall.Close(watcher.fd)
}
