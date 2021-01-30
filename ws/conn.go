package ws

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// Connection status
const (
	Established = 1
	Closing     = 2
	Closed      = 3
)

// Conn represents a WebSocket connection
type Conn struct {
	Rwc   net.Conn
	r     *bufio.Reader
	State int
}

// NewConn is a constructor of Conn
func NewConn(tcpConn net.Conn) *Conn {
	return &Conn{tcpConn, bufio.NewReader(tcpConn), Established}
}

// ReadMessage read received frames and return DataFrame object
func (c *Conn) ReadMessage() (*DataFrame, error) {
	df, err := NewDataFrameFromReader(c.r)
	if err != nil {
		return nil, err
	}
	return df, nil
}

// ReadTextMessage handles text message from client
func (c *Conn) ReadTextMessage() (string, error) {
	df, err := c.ReadMessage()
	if err != nil {
		return "", err
	}
	if df.OpCode != OpCodeText {
		return "", fmt.Errorf("Invalid format")
	}
	return string(df.Message()), nil
}

// ReadBinaryMessage handles binary message from client
func (c *Conn) ReadBinaryMessage() ([]byte, error) {
	df, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	if df.OpCode != OpCodeBinary {
		return nil, fmt.Errorf("Invalid format")
	}
	return df.Message(), nil
}

// SendTextMessage pushes text message to client
func (c *Conn) SendTextMessage(msg string) error {
	df, err := NewDataFrameFromTextMessage(msg, false)
	if err != nil {
		return err
	}
	_, err = c.Rwc.Write(df.Frame())
	if err != nil {
		return err
	}
	return nil
}

// SendBinaryMessage pushes binary message to client
func (c *Conn) SendBinaryMessage(msg []byte) error {
	df, err := NewDataFrameFromBinaryMessage(msg, false)
	if err != nil {
		return err
	}
	_, err = c.Rwc.Write(df.Frame())
	if err != nil {
		return err
	}
	return nil
}

// SendCloseFrame send an close frame
func (c *Conn) SendCloseFrame(status int) error {
	// status code must be network byte order
	// https://tools.ietf.org/html/rfc6455#section-5.5.1
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(status))
	df, err := NewDataFrameFromBinaryMessage(buf, false)
	if err != nil {
		return err
	}
	df.OpCode = OpCodeClose
	_, err = c.Rwc.Write(df.Frame())
	if err != nil {
		return err
	}
	return nil
}

// Close closes connection
func (c *Conn) Close() {
	if c.State == Established {
		c.SendCloseFrame(StatusNormalClosure)
		c.State = Closing
		return
	}
	c.Rwc.Close()
	c.State = Closed
	return
}

// Ping sends a ping to client
func (c *Conn) Ping() error {
	msg := fmt.Sprintf("ping %d", time.Now().Unix())
	df, err := NewDataFrameFromTextMessage(msg, false)
	if err != nil {
		return err
	}
	df.OpCode = OpCodePing
	_, err = c.Rwc.Write(df.Frame())
	if err != nil {
		return err
	}
	return nil
}
