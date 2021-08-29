package impl

import (
	"net"
)

type Source struct {
	net.Conn
	pid       uint32
	secretKey uint32
}

func (s *Source) GetID() uint32 {
	return s.pid
}
