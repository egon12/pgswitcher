package impl

import (
	"errors"
	"io"

	"github.com/egon12/pgswitcher/ctrl"
	"github.com/jackc/pgproto3/v2"
)

// Piper implementaiton of ctrl.Piper
type Piper struct {
	b  *buffer
	tr *TargetReader
}

// this function should be block the process
// and will return when source starting the chat
func (p *Piper) WaitForChat(s ctrl.Source) error {
	p.b.Reset()
	err := p.b.ReadFrom(s)
	if err != nil {
		return err
	}

	// let's make terminate command into EOF
	if p.b.lastRead > 0 && p.b.buf[0] == 'X' {
		return io.EOF
	}

	return nil
}

// Chat is send message form source to target and got the
// result
func (p *Piper) Chat(s ctrl.Source, t ctrl.Target) error {
	var err error

	if !p.b.HaveMessage() {
		return errors.New("source doesn't start a chat")
	}

	err = p.b.WriteTo(t)
	if err != nil {
		return err
	}

	p.tr.SetTarget(t)

	err = p.tr.ReadTo(s)
	if err != nil {
		return err
	}

	for p.tr.StillChat() {
		err = p.b.ReadFrom(s)
		if err != nil {
			return err
		}

		err = p.b.WriteTo(t)
		if err != nil {
			return err
		}

		err = p.tr.ReadTo(s)
		if err != nil {
			return err
		}
	}

	return nil

}

// buffer is some a copy from bytes.buffer with different
// behaviour
type buffer struct {
	buf      []byte
	lastRead int
}

func newBuffer() *buffer {
	b := &buffer{}
	b.buf = make([]byte, bufferSize)
	b.lastRead = 0
	return b
}

func (b *buffer) Reset() {
	b.buf = make([]byte, bufferSize)
	b.lastRead = 0
}

func (b *buffer) ReadFrom(r io.Reader) error {
	n, err := r.Read(b.buf)
	if err != nil {
		return err
	}

	if n < 0 {
		return errors.New("negative from read")
	}

	if n == bufferSize {
		return errors.New("query must less than 10 MB")
	}

	b.lastRead = n

	return nil
}

func (b *buffer) HaveMessage() bool {
	return b.lastRead > 0

}

func (b *buffer) WriteTo(w io.Writer) error {
	n, err := w.Write(b.buf[0:b.lastRead])
	if n != b.lastRead {
		return errors.New("failed to send full message")
	}
	return err
}

func (b *buffer) String() string {
	return string(b.buf[0:b.lastRead])
}

// TargetReader
type TargetReader struct {
	stillChat bool
	frontend  *pgproto3.Frontend
}

func (tr *TargetReader) SetTarget(t ctrl.Target) {
	tr.frontend = pgproto3.NewFrontend(
		pgproto3.NewChunkReader(t),
		t,
	)
}

func (tr *TargetReader) ReadTo(s ctrl.Source) error {
	var err error

	bm, err := tr.frontend.Receive()
	if err != nil {
		return err
	}

	_, err = s.Write(bm.Encode(nil))
	if err != nil {
		return err
	}

	for tr.notReadyForQuery(bm) {
		bm, err = tr.frontend.Receive()
		if err != nil {
			return err
		}

		_, err = s.Write(bm.Encode(nil))
		if err != nil {
			return err
		}
	}

	return nil
}

func (tr *TargetReader) StillChat() bool { return tr.stillChat }

func (tr *TargetReader) notReadyForQuery(m pgproto3.BackendMessage) bool {
	//fmt.Printf("got %#v\n", m)
	switch m.(type) {
	default:
		return true
	case *pgproto3.ReadyForQuery:
		z := m.(*pgproto3.ReadyForQuery)
		// I for Idl
		// T for transaction
		// E for failed transaction

		tr.stillChat = z.TxStatus != 'I'

		return false
	}
}

const bufferSize = 10 * 1024 * 1024 // 10 MB / connection
