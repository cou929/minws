package ws

import (
	"bufio"
	"fmt"
	"log"
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
	buf := make([]byte, 2)
	n, err := c.r.Read(buf)
	if err != nil {
		return "", err
	}
	if n != 2 {
		log.Println("read bytes", n)
		return "", fmt.Errorf("invalid frame format %#b", n)
	}

	fin := buf[0] >> 7
	opCode := buf[0] & 0b00001111
	mask := buf[1] >> 7
	payloadLen := int(buf[1] & 0b01111111)

	log.Printf("read %d byte, %v %b", n, buf, buf)
	log.Println("fin", fin, "opCode", opCode, "mask", mask, "payloadLen", payloadLen)

	if payloadLen >= 126 {
		return "", fmt.Errorf("extended payload length is not supported yet")
	}

	maskingKey := make([]byte, 4)
	n, err = c.r.Read(maskingKey)
	if err != nil {
		return "", err
	}
	if n != 4 {
		log.Println("read bytes", n)
		return "", fmt.Errorf("invalid frame format %#b", n)
	}

	log.Println("maskingKey", maskingKey)

	encoded := make([]byte, payloadLen)
	n, err = c.r.Read(encoded)
	if err != nil {
		return "", err
	}
	if n != payloadLen {
		log.Println("read bytes", n)
		return "", fmt.Errorf("invalid frame format %#b", n)
	}

	log.Println("encoded payload", encoded)

	decoded := make([]byte, payloadLen)
	for i := 0; i < payloadLen; i++ {
		decoded[i] = encoded[i] ^ maskingKey[i%4]
	}

	log.Printf("decoded %s\n", string(decoded))

	return string(decoded), nil
}

// SendMessage pushes message to client
func (c *Conn) SendMessage(msg string) error {
	bit := int8(len(msg))
	bit = bit & 0b01111111 // mask bit off
	message := []byte{0b10000001, byte(bit)}
	message = append(message, ([]byte)(msg)...)
	n, err := c.Rwc.Write(message)
	if err != nil {
		return err
	}
	log.Println("sent byte", n, "message len", len(message))
	return nil
}

// Close closes connection
func (c *Conn) Close() {
	return
}
