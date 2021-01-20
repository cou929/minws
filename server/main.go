package main

import (
	"log"
	"net"

	minhttp "github.com/cou929/minws/server/http"
)

func main() {
	srv := NewServer(":5001")
	log.Fatal(srv.Serve())
}

// Server represents server of minws
type Server struct {
	address string
	http    *minhttp.Server
}

// NewServer is constructor of Server
func NewServer(address string) *Server {
	return &Server{
		address: address,
		http:    minhttp.NewServer(),
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

func (srv *Server) serve(c *Conn) error {
	defer func() {
		c.rwc.Close()
	}()
	for {
		switch c.status {
		case Initialized:
			srv.http.HandShake(c.rwc)
		}
	}
}

// Conn represents a connection
type Conn struct {
	rwc    net.Conn
	status ConnStatus
}

func newConn(c net.Conn) *Conn {
	return &Conn{
		rwc:    c,
		status: Initialized,
	}
}

// ConnStatus represents a status of the connection
type ConnStatus int

const (
	// Initialized represents connection status which is not hand-shacked yet
	Initialized ConnStatus = iota + 1
	// HandShaked represents hand-shacked connection status
	HandShaked
)
