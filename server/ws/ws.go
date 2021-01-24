package ws

import (
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
	log.Println("todo")
	conn.Rwc.Close()
	return nil
}
