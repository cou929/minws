package main

import (
	"log"
	"net"

	"github.com/cou929/minws"
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
				msg, err := c.ReadMessage()
				if err != nil {
					log.Println(err)
					c.Close()
					return
				}
				log.Println("on message", msg)
				err = c.SendMessage("Hello World!")
				if err != nil {
					log.Println(err)
					c.Close()
					return
				}
			}
		}(conn)
	}
}
