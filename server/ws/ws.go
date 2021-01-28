package ws

import (
	"bufio"
	"fmt"
	"log"

	minwsconn "github.com/cou929/minws/server/conn"
)

// Server represents server to handle WebSocket
type Server struct{}

// NewServer is constructor of Server
func NewServer() *Server {
	return &Server{}
}

// HandleMessage handles WebSocket message
func (srv *Server) HandleMessage(conn *minwsconn.Conn) error {
	for {
		r := bufio.NewReader(conn.Rwc)
		for {
			buf := make([]byte, 2)
			n, err := r.Read(buf)
			if err != nil {
				return err
			}
			if n != 2 {
				log.Println("read bytes", n)
				return fmt.Errorf("invalid frame format %#b", n)
			}

			fin := buf[0] >> 7
			opCode := buf[0] & 0b00001111
			mask := buf[1] >> 7
			payloadLen := int(buf[1] & 0b01111111)

			log.Printf("read %d byte, %v %b", n, buf, buf)
			log.Println("fin", fin, "opCode", opCode, "mask", mask, "payloadLen", payloadLen)

			if payloadLen >= 126 {
				return fmt.Errorf("extended payload length is not supported yet")
			}

			maskingKey := make([]byte, 4)
			n, err = r.Read(maskingKey)
			if err != nil {
				return err
			}
			if n != 4 {
				log.Println("read bytes", n)
				return fmt.Errorf("invalid frame format %#b", n)
			}

			log.Println("maskingKey", maskingKey)

			encoded := make([]byte, payloadLen)
			n, err = r.Read(encoded)
			if err != nil {
				return err
			}
			if n != payloadLen {
				log.Println("read bytes", n)
				return fmt.Errorf("invalid frame format %#b", n)
			}

			log.Println("encoded payload", encoded)

			decoded := make([]byte, payloadLen)
			for i := 0; i < payloadLen; i++ {
				decoded[i] = encoded[i] ^ maskingKey[i%4]
			}

			log.Printf("decoded %s\n", string(decoded))

			// send
			payload := "hello world!!"
			bit := int8(len(payload))
			bit = bit & 0b01111111 // mask bit off
			message := []byte{0b10000001, byte(bit)}
			message = append(message, ([]byte)(payload)...)
			n, err = conn.Rwc.Write(message)
			if err != nil {
				return err
			}
			log.Println("sent byte", n, "message len", len(message))
		}
	}
	return nil
}
