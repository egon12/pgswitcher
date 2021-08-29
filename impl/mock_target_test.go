package impl

import (
	"bytes"

	"github.com/jackc/pgproto3/v2"
)

type targetMock struct {
	w        *bytes.Buffer
	r        *bytes.Buffer
	released bool
}

func newTargetMock() *targetMock {
	return &targetMock{&bytes.Buffer{}, &bytes.Buffer{}, false}
}

func (t *targetMock) SetBackendMessage(m pgproto3.BackendMessage) {
	t.r.Write(m.Encode(nil))
}

func (t *targetMock) Write(b []byte) (int, error) { return t.w.Write(b) }
func (t *targetMock) Read(b []byte) (int, error)  { return t.r.Read(b) }
func (t *targetMock) Release()                    { t.released = true }
