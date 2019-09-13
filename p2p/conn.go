// Copy of https://github.com/libp2p/go-libp2p-gostream/blob/master/conn.go
package p2p

import (
	"net"
	"time"

	"github.com/libp2p/go-libp2p-core/helpers"
	"github.com/libp2p/go-libp2p-core/network"
)

// conn is an implementation of net.Conn which wraps
// libp2p streams.
type conn struct {
	s network.Stream
}

// NewNetConn creates a conn given a libp2p stream
func NewNetConn(s network.Stream) net.Conn {
	return &conn{s}
}

// Read reads data from the connection.
func (c *conn) Read(b []byte) (n int, err error) {
	return c.s.Read(b)
}

// Write writes data to the connection.
func (c *conn) Write(b []byte) (n int, err error) {
	return c.s.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *conn) Close() error {
	if err := c.s.Close(); err != nil {
		_ = c.s.Reset()
		return err
	}
	go func() {
		_ = helpers.AwaitEOF(c.s)
	}()
	return nil
}

// LocalAddr returns the local network address.
func (c *conn) LocalAddr() net.Addr {
	return &addr{c.s.Conn().LocalPeer()}
}

// RemoteAddr returns the remote network address.
func (c *conn) RemoteAddr() net.Addr {
	return &addr{c.s.Conn().RemotePeer()}
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
// See https://golang.org/pkg/net/#Conn for more details.
func (c *conn) SetDeadline(t time.Time) error {
	return c.s.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (c *conn) SetReadDeadline(t time.Time) error {
	return c.s.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.s.SetWriteDeadline(t)
}
