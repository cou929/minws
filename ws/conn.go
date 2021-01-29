package ws

import (
	"bufio"
	"net"
)

// Conn represents a WebSocket connection
type Conn struct {
	Rwc net.Conn
	r   *bufio.Reader
}

// NewConn is a constructor of Conn
func NewConn(tcpConn net.Conn) *Conn {
	return &Conn{tcpConn, bufio.NewReader(tcpConn)}
}

// ReadMessage handles message from client
func (c *Conn) ReadMessage() (string, error) {
	df, err := NewDataFrameFromReader(c.r)
	if err != nil {
		return "", err
	}
	return df.Message(), nil
}

// SendMessage pushes message to client
func (c *Conn) SendMessage(msg string) error {
	df, err := NewDataFrameFromMessage(msg, false)
	if err != nil {
		return err
	}
	_, err = c.Rwc.Write(df.Frame())
	if err != nil {
		return err
	}
	return nil
}

// Close closes connection
func (c *Conn) Close() {
	return
}
