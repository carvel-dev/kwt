package dstconn

import (
	"io"
	"net"
	"sync"
)

type SSHConnCopier struct {
	logTag string
	logger Logger
}

var _ ConnCopier = SSHConnCopier{}

type sshConnSrc interface {
	io.Reader
	io.Writer
	CloseWrite() error
	// CloseRead() error // see below
	Close() error
}

type sshConnDst interface {
	io.Reader
	io.Writer
	CloseWrite() error
	Close() error
}

func NewSSHConnCopier(logTagSuffix string, logger Logger) SSHConnCopier {
	return SSHConnCopier{"SSHConnCopier " + logTagSuffix, logger}
}

func (c SSHConnCopier) CopyAndClose(dstConn, srcConn net.Conn) {
	dstConnCloser := dstConn.(sshConnDst)
	srcConnCloser := srcConn.(sshConnSrc)

	var wg sync.WaitGroup

	wg.Add(2)

	go c.srcToDstCopy(dstConnCloser, srcConnCloser, &wg)
	go c.dstToSrcCopy(dstConnCloser, srcConnCloser, &wg)

	wg.Wait()

	err := srcConn.Close()
	if err != nil {
		if err != io.EOF {
			c.logger.Error(c.logTag, "Failed to close src conn: %s", err)
		}
	}

	err = dstConn.Close()
	if err != nil {
		// TODO somehwat weird (if CloseWrite is called before, Close will return EOF)
		if err != io.EOF {
			c.logger.Error(c.logTag, "Failed to close dst conn: %s", err)
		}
	}
}

func (c SSHConnCopier) srcToDstCopy(dstConn sshConnDst, srcConn sshConnSrc, wg *sync.WaitGroup) {
	_, err := io.Copy(dstConn, srcConn)
	if err != nil {
		c.logger.Error(c.logTag, "Failed to copy src->dst conn: %s", err)
	}

	// TODO gracefully end reading?
	// err = srcConn.CloseRead()
	// if err != nil {
	// 	c.logger.Error(c.logTag, "Failed to close-read src conn: %s", err)
	// }

	err = dstConn.CloseWrite()
	if err != nil {
		c.logger.Error(c.logTag, "Failed to close-write dst conn: %s", err)
	}

	wg.Done()
}

func (c SSHConnCopier) dstToSrcCopy(dstConn sshConnDst, srcConn sshConnSrc, wg *sync.WaitGroup) {
	_, err := io.Copy(srcConn, dstConn)
	if err != nil {
		c.logger.Error(c.logTag, "Failed to copy dst->src conn: %s", err)
	}

	err = srcConn.CloseWrite()
	if err != nil {
		c.logger.Error(c.logTag, "Failed to close-write src conn: %s", err)
	}

	wg.Done()
}
