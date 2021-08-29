package impl

import (
	"net"
	"testing"

	"github.com/jackc/pgproto3/v2"
)

func TestAsdfg(t *testing.T) {
	//client, server := net.Pipe()

	//go func() {
	//	NewSource(server, &TargetPool{})

	//	server.Close()
	//}()

	client, _ := net.Dial("tcp", "127.0.0.1:5432")

	f := pgproto3.NewFrontend(pgproto3.NewChunkReader(client), client)

	sm := &pgproto3.StartupMessage{
		ProtocolVersion: pgproto3.ProtocolVersionNumber,
		Parameters: map[string]string{
			"user":             "nakama",
			"database":         "nakama",
			"application_name": "something",
			"client_encoding":  "UTF-8",
		},
	}

	err := f.Send(sm)
	if err != nil {
		t.Error(err)
	}

	bm, err := f.Receive()
	t.Errorf("%#v", bm)
	t.Error(err)

	client.Close()
}
