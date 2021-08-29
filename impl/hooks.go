package impl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/egon12/pgswitcher/ctrl"
	"github.com/jackc/pgproto3/v2"
)

type Hooks struct {
	sqlfilepath string
	sqlfile     *os.File
	sqlinput    io.Reader
}

func NewHooks(sqlfilepath string) (*Hooks, error) {
	file, err := os.Open(sqlfilepath)
	if err != nil {
		return nil, err
	}

	return &Hooks{
		sqlfilepath: sqlfilepath,
		sqlfile:     file,
		sqlinput:    file,
	}, nil

}

func (h *Hooks) BeforeUseNew(o ctrl.Target, new ctrl.Target) error {
	var err error

	fe := pgproto3.NewFrontend(pgproto3.NewChunkReader(new), new)
	q := &pgproto3.Query{}

	scanner := bufio.NewScanner(h.sqlinput)
	for scanner.Scan() {
		q.String = scanner.Text()

		if q.String == "" {
			continue
		}

		err = fe.Send(q)
		if err != nil {
			return fmt.Errorf("Send(query) failed: %v", err)
		}

		bm, err := fe.Receive()
		if err != nil {
			return fmt.Errorf("Receive(query) failed: %v", err)
		}

		for h.notReady(bm) {
			err = h.parseError(bm)
			if err != nil {
				return err
			}
			bm, err = fe.Receive()
			if err != nil {
				return fmt.Errorf("Receive(query) failed: %v", err)
			}
		}

	}
	return nil
}

func (h *Hooks) notReady(bm pgproto3.BackendMessage) bool {
	switch bm.(type) {
	default:
		return true
	case *pgproto3.ReadyForQuery:
		return false
	}
}

func (h *Hooks) parseError(bm pgproto3.BackendMessage) error {
	switch bm.(type) {
	default:
		return nil
	case *pgproto3.ErrorResponse:
		er := bm.(*pgproto3.ErrorResponse)
		return errors.New(er.Message)
	}
}
