package pf

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// https://github.com/apple/darwin-xnu/blob/0a798f6738bc1db01281fc08ae024145e84df927/bsd/sys/ioccom.h#L91
const IOCPARM_MASK uint32 = 0x1fff
const IOC_OUT uint32 = 0x40000000
const IOC_IN uint32 = 0x80000000
const IOC_INOUT uint32 = IOC_IN | IOC_OUT
const IOC_GROUP_D = 68 // 'D'

func _IOC(inout, group, num, len uint32) uint32 {
	return inout | ((len & IOCPARM_MASK) << 16) | (group << 8) | num
}

type Ioctl struct {
	f *os.File
}

func NewIoctl(path string) (*Ioctl, error) {
	f, err := os.OpenFile(path, syscall.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	return &Ioctl{f}, nil
}

func (c *Ioctl) Close() error {
	return c.f.Close()
}

func (c *Ioctl) Read(cmd uint32, ptr unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(c.f.Fd()), uintptr(cmd), uintptr(ptr))
	if errno != 0 {
		return fmt.Errorf("ioctl error: %s", errno)
	}

	return nil
}
