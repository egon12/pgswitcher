package impl

import (
	"net"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Target struct {
	net.Conn
	pool     *TargetPool
	poolConn *pgxpool.Conn
}

func (t *Target) Release() {
	t.pool.Release(t)
}
