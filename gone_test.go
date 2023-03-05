package gone_test

import (
	"fmt"
	"log"
	"syscall"
	"testing"
	"unsafe"
)

func TestMain(m *testing.M) {
	fd, err := syscall.InotifyInit()
	if err != nil {
		log.Fatal(err.Error())
	}
	path := "./"
	_, err = syscall.InotifyAddWatch(fd, path, syscall.IN_MODIFY)
	if err != nil {
		syscall.Close(fd)
		log.Fatal(err.Error())
	}

	events := make(chan uint32)
	go func() {
		var buf [syscall.SizeofInotifyEvent * 4096]byte
		for {
			n, err := syscall.Read(fd, buf[:])
			if err != nil {
				n = 0
				continue
			}
			var offset uint32
			for offset <= uint32(n-syscall.SizeofInotifyEvent) {
				raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
				mask := uint32(raw.Mask)
				nameLen := uint32(raw.Len)
				events <- mask
				offset += syscall.SizeofInotifyEvent + nameLen
			}
		}
	}()

	for {
		select {
		case event := <-events:
			if event&syscall.IN_MODIFY == syscall.IN_MODIFY {
				fmt.Println("asd")
			}
		}
	}
}
