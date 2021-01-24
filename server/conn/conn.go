package conn

import "net"

// Conn represents a connection
type Conn struct {
	Rwc    net.Conn
	Status Status
}

// CompleteHandShake transits conn status to HandShaked
func (c *Conn) CompleteHandShake() {
	c.Status = HandShaked
}

// Status represents a status of the connection
type Status int

const (
	// Initialized represents connection status which is not hand-shacked yet
	Initialized Status = iota + 1
	// HandShaked represents hand-shacked connection status
	HandShaked
)
