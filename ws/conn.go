package ws

import (
	"bufio"
	"fmt"
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

// Close closes connection
func (c *Conn) Close() {
	return
}
