package impl

import (
	"net"

	"github.com/egon12/pgswitcher/ctrl"
	"github.com/jackc/pgconn"
)

type SourceFactory struct {
	fb *fakeBackend
}

func NewSourceFactory(urls string, h *pgconn.HijackedConn) (*SourceFactory, error) {
	fb, err := newFakeBackend(urls)
	if err != nil {
		return nil, err
	}

	fb.setServerParameter(h)

	return &SourceFactory{
		fb: fb,
	}, nil
}

func (sf *SourceFactory) New(c net.Conn) (ctrl.Source, error) {
	pid, secretKey, err := sf.fb.handleLogin(c)
	if err != nil {
		return nil, err
	}

	return &Source{
		Conn:      c,
		pid:       pid,
		secretKey: secretKey,
	}, nil
}
