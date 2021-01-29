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
				default:
					log.Println("not message", df.OpCode)
				}
			}
		}(conn)
	}
}
