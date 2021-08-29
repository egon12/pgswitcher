package impl

import (
	"testing"

	"github.com/jackc/pgproto3/v2"
)

func TestHooks(t *testing.T) {
	hooks, err := NewHooks("hooks_test.sql")
	if err != nil {
		t.Fatal(err)
	}

	o := newTargetMock()
	n := newTargetMock()

	n.SetBackendMessage(&pgproto3.ReadyForQuery{'I'})

	err = hooks.BeforeUseNew(o, n)
	if err != nil {
		t.Error(err)
	}

	b := n.w.Bytes()

	sqlb := b[5:]
	want := []byte("SELECT setval('table_01_id_seq', (SELECT max(id) FROM table_01))")
	// there are 0 byte in the end of query
	want = append(want, 0)
	if string(sqlb) != string(want) {
		t.Errorf("got wrong sql \n%x\n%x", sqlb, want)
	}
}
