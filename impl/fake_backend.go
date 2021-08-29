package impl

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync/atomic"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
)

type fakeBackend struct {
	be         *pgproto3.Backend
	allowedURL string
	url        *url.URL
	urls       []*url.URL

	parameterStatuses []*pgproto3.ParameterStatus

	pid       uint32
	secretKey uint32
}

var (
	requestPassword = &pgproto3.AuthenticationCleartextPassword{}

	readyForQuery = &pgproto3.ReadyForQuery{'I'}

	authOK = &pgproto3.AuthenticationOk{}
)

func newFakeBackend(allowedURL string) (*fakeBackend, error) {
	u, err := url.Parse(allowedURL)
	if err != nil {
		return nil, err
	}

	return &fakeBackend{
		url:       u,
		pid:       1,
		secretKey: 1,
	}, nil
}

func (fb *fakeBackend) handleLogin(c net.Conn) (pid, secretKey uint32, err error) {
	fb.setBackend(c)

	defer func() {
		if err != nil {
			fb.sendError("FATAL", err)
		}
	}()

	m, err := fb.be.ReceiveStartupMessage()
	if err != nil {
		err = fmt.Errorf("ReceiveStartupMessage failed: %w", err)
		return
	}

	sm, err := fb.ensureSM(m)
	if err != nil {
		return
	}

	err = fb.checkSM(sm)
	if err != nil {
		return
	}

	err = fb.be.Send(requestPassword)
	if err != nil {
		err = fmt.Errorf("Send(requestPassword) failed: %w", err)
		return
	}

	m, err = fb.be.Receive()
	if err != nil {
		err = fmt.Errorf("Receive PasswordMessage failed: %w", err)
		return
	}

	pm, err := fb.ensurePM(m)
	if err != nil {
		return
	}

	err = fb.checkPM(pm)
	if err != nil {
		return
	}

	for _, v := range fb.parameterStatuses {
		err = fb.be.Send(v)
		if err != nil {
			err = fmt.Errorf("Send(parameterStatus) failed: %w", err)
			return
		}
	}

	err = fb.be.Send(&pgproto3.BackendKeyData{
		ProcessID: fb.pid,
		SecretKey: fb.secretKey,
	})
	if err != nil {
		err = fmt.Errorf("Send(backendKeyData) failed: %w", err)
		return
	}

	pid = atomic.AddUint32(&fb.pid, 1)
	secretKey = atomic.AddUint32(&fb.secretKey, 1)

	err = fb.be.Send(readyForQuery)
	if err != nil {
		err = fmt.Errorf("Send(readyForQuery) failed: %w", err)
		return
	}

	return
}

func (fb *fakeBackend) checkSM(m *pgproto3.StartupMessage) error {
	if fb.url.User.Username() != m.Parameters["user"] {
		return fmt.Errorf(`password authentication failed for user "%s"`, m.Parameters["user"])
	}

	if fb.url.Path[1:] != m.Parameters["database"] {
		return fmt.Errorf("got wrong database: \"%s\"", m.Parameters["database"])
	}

	//if "UTF8" != m.Parameters["client_encoding"] {
	//	return fmt.Errorf("got wrong client_encoding: \"%s\"",
	//		m.Parameters["client_encoding"])
	//}

	return nil
}

func (fb *fakeBackend) checkPM(m *pgproto3.PasswordMessage) error {
	password, ok := fb.url.User.Password()
	if !ok {
		return nil
	}

	if m.Password != password {
		return errors.New("got wrong password")
	}

	return fb.be.Send(authOK)
}

func (fb *fakeBackend) ensureSM(m pgproto3.FrontendMessage) (*pgproto3.StartupMessage, error) {
	switch m.(type) {
	default:
		return nil, fmt.Errorf("expect *pgproto3.StartupMessage got %T", m)
	case *pgproto3.StartupMessage:
		return m.(*pgproto3.StartupMessage), nil
	}
}

func (fb *fakeBackend) ensurePM(m pgproto3.FrontendMessage) (*pgproto3.PasswordMessage, error) {
	switch m.(type) {
	default:
		return nil, fmt.Errorf("expect *pgproto3.PasswordMessage got %T", m)
	case *pgproto3.PasswordMessage:
		return m.(*pgproto3.PasswordMessage), nil
	}
}

func (fb *fakeBackend) sendError(severity string, err error) {
	_ = fb.be.Send(&pgproto3.ErrorResponse{
		Severity: severity,
		Message:  err.Error(),
	})
}

func (fb *fakeBackend) setBackend(conn net.Conn) {
	fb.be = pgproto3.NewBackend(
		pgproto3.NewChunkReader(conn),
		conn,
	)
}

func (fb *fakeBackend) setServerParameter(c *pgconn.HijackedConn) {
	fb.parameterStatuses = make([]*pgproto3.ParameterStatus, len(c.ParameterStatuses))
	i := 0
	for n, v := range c.ParameterStatuses {
		fb.parameterStatuses[i] = &pgproto3.ParameterStatus{Name: n, Value: v}
		i++
	}
}
