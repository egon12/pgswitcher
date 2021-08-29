package ctrl

import (
	"bytes"
	"errors"
	"net"
	"testing"
)

func TestNewController(t *testing.T) {
	c := NewController(nil, nil, nil, nil, nil)
	if c == nil {
		t.Error("controller should not be nil")
	}
}

func TestNewController_Handle(t *testing.T) {
	var sf sourceFactoryMock = func(net.Conn) (Source, error) {
		return &sourceMock{&bytes.Buffer{}, 1}, nil
	}

	tp := &targetPoolMock{false, &targetMock{}, &targetMock{}}

	pf := &piperFactoryMock{nil, errors.New("ex")}

	c := NewController(sf, tp, pf, nil, nil)

	_, server := net.Pipe()

	c.Handle(server)
}

func TestNewController_stream(t *testing.T) {
	tp := &targetPoolMock{false, &targetMock{}, &targetMock{}}

	pf := &piperFactoryMock{nil, errors.New("ex")}

	c := NewController(nil, tp, pf, nil, nil)

	c.stream(newSourceMock(0))
}

type sourceFactoryMock func(net.Conn) (Source, error)

func (s sourceFactoryMock) New(c net.Conn) (Source, error) { return s(c) }

type sourceMock struct {
	*bytes.Buffer
	id uint32
}

func newSourceMock(id uint32) *sourceMock {
	return &sourceMock{&bytes.Buffer{}, id}
}

func (s *sourceMock) GetID() uint32 { return s.id }

type targetMock struct {
	*bytes.Buffer
	released bool
}

func (s *targetMock) Release() { s.released = true }

type targetPoolMock struct {
	useNew     bool
	oldTargets *targetMock
	newTargets *targetMock
}

func (t *targetPoolMock) Switch(new bool) error       { t.useNew = new; return nil }
func (t *targetPoolMock) UseNew() bool                { return t.useNew }
func (t *targetPoolMock) Acquire() (Target, error)    { return t.oldTargets, nil }
func (t *targetPoolMock) AcquireNew() (Target, error) { return t.newTargets, nil }
func (t *targetPoolMock) Close() error                { panic("not implemented") }

type piperMock struct{ w, c error }

func (p *piperMock) WaitForChat(_ Source) error    { return p.w }
func (p *piperMock) Chat(_ Source, _ Target) error { return p.c }

type piperFactoryMock struct{ w, c error }

func (p *piperFactoryMock) New() Piper { return &piperMock{p.w, p.c} }
