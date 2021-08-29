package impl

import (
	"net"
	"os"
	"testing"

	"github.com/jackc/pgproto3/v2"
)

func TestFakeBackend_handleLogin_real(t *testing.T) {
	addr := os.Getenv("PSQL")
	if len(addr) == 0 {
		t.Skip("for real test by psql")
	}

	l, _ := net.Listen("tcp", ":8080")
	s, _ := l.Accept()

	fb, _ := newFakeBackend(addr)

	fb.handleLogin(s)
}

func TestFakeBackend_handleLogin(t *testing.T) {
	c, s := net.Pipe()

	fb, _ := newFakeBackend(
		"postgres://user:pass@127.0.0.1/db?sslmode=disable",
	)
	go func() {
		_, _, err := fb.handleLogin(s)
		if err != nil {
			t.Error(err)
		}
	}()

	f := pgproto3.NewFrontend(pgproto3.NewChunkReader(c), c)

	f.Send(&pgproto3.StartupMessage{
		ProtocolVersion: pgproto3.ProtocolVersionNumber,
		Parameters: map[string]string{
			"user":             "user",
			"database":         "db",
			"application_name": "app",
			"client_encoding":  "UTF8",
		},
	})

	// receive password request
	_, _ = f.Receive()

	f.Send(&pgproto3.PasswordMessage{"pass"})

	// receive authentication ok
	_, _ = f.Receive()

	// receive authentication ready for query
	_, _ = f.Receive()
}

func TestFakeBackend_handleLogin_failed_wrong_sm(t *testing.T) {

	c, s := net.Pipe()

	fb, _ := newFakeBackend(
		"postgres://user:pass@127.0.0.1/db?sslmode=disable",
	)

	go fb.handleLogin(s)

	msg := &pgproto3.StartupMessage{
		ProtocolVersion: pgproto3.ProtocolVersionNumber,
		Parameters: map[string]string{
			"user":             "user",
			"database":         "db",
			"application_name": "app",
			"client_encoding":  "UTF8",
		},
	}

	f := pgproto3.NewFrontend(pgproto3.NewChunkReader(c), c)

	msg.Parameters["user"] = "user1"

	f.Send(msg)

	// receive password request
	m, _ := f.Receive()

	e := m.(*pgproto3.ErrorResponse)
	if e.Message != `password authentication failed for user "user1"` {
		t.Errorf("got wrong error message: %s", e.Message)
	}

}
