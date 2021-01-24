package server

import (
	"log"
	"net"

	"github.com/cou929/minws/server/conn"
	minhttp "github.com/cou929/minws/server/http"
	minws "github.com/cou929/minws/server/ws"
)

// Server represents server of minws
type Server struct {
	address string
	http    *minhttp.Server
	ws      *minws.Server
}

// NewServer is constructor of Server
func NewServer(address string) *Server {
	return &Server{
		address: address,
		http:    minhttp.NewServer(),
		ws:      minws.NewServer(),
	}
}

// Serve HTTP and will WebSocket
func (srv *Server) Serve() error {
	l, err := net.Listen("tcp", srv.address)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		c := newConn(conn)
		go srv.serve(c)
	}
}

func (srv *Server) serve(c *conn.Conn) error {
	defer func() {
		c.Rwc.Close()
	}()
	for {
		var err error
		switch c.Status {
		case conn.Initialized:
			err = srv.http.HandShake(c)
		case conn.HandShaked:
			err = srv.ws.HandleMessage(c)
		}
		if err != nil {
			log.Println(err)
		}
	}
}

func newConn(c net.Conn) *conn.Conn {
	return &conn.Conn{
		Rwc:    c,
		Status: conn.Initialized,
	}
}
