package main

import (
	"log"
	"net"

	"github.com/cou929/minws"
	"github.com/cou929/minws/ws"
)

func main() {
	l, err := net.Listen("tcp", ":5001")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(tcpConn net.Conn) {
			c, err := minws.HandShake(tcpConn)
			if err != nil {
				log.Println(err)
				c.Close()
				return
			}
			defer c.Close()
			for {
				df, err := c.ReadMessage()
				if err != nil {
					log.Println(err)
					c.Close()
					return
				}
				switch df.OpCode {
				case ws.OpCodeText:
					msg := string(df.Message())
					log.Println("on text message", msg)
					err = c.SendTextMessage("echoed: " + msg)
					if err != nil {
						log.Println(err)
						c.Close()
						return
					}
				case ws.OpCodeBinary:
					msg := df.Message()
					log.Println("on binary message", msg)
					msg = append([]byte{1}, msg...)
					err = c.SendBinaryMessage(msg)
					if err != nil {
						log.Println(err)
						c.Close()
						return
					}
					c.Close()
				case ws.OpCodeClose:
					status, err := (df.CloseStatusCode())
					if err != nil {
						log.Println(err)
						c.Close()
						return
					}
					log.Println("on close", status, ws.StatusText(status))
					switch c.State {
					case ws.Established:
						err = c.SendCloseFrame(status)
						if err != nil {
							log.Println(err)
							c.Close()
							return
						}
						c.Close()
						return
					case ws.Closing:
						c.Close()
						return
					case ws.Closed:
						log.Println("received close frame on closed state conn")
						c.Close()
						return
					}
				case ws.OpCodePing:
					msg := df.Message()
					log.Println("on ping", msg)
					err := c.SendBinaryMessage(msg)
					if err != nil {
						log.Println(err)
						c.Close()
						return
					}
				case ws.OpCodePong:
					msg := df.Message()
					log.Println("on pong", string(msg))
				default:
					log.Println("not message", df.OpCode)
				}
			}
		}(conn)
	}
}
