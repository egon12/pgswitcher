package impl

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/egon12/pgswitcher/ctrl"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TargetPool struct {
	o      *pgxpool.Pool
	n      *pgxpool.Pool
	useNew bool
	lock   *sync.Mutex
}

func NewTargetPool(oldurl, newurl string, useNew bool) (*TargetPool, error) {
	o, err := pgxpool.Connect(context.Background(), oldurl)
	if err != nil {
		return nil, err
	}

	n, err := pgxpool.Connect(context.Background(), newurl)
	if err != nil {
		return nil, err
	}

	return &TargetPool{
		o:      o,
		n:      n,
		useNew: useNew,
		lock:   &sync.Mutex{},
	}, nil
}

// Switch will switch the connection from old connection to
// new connection
// If we still have old connection alive until timeout, it should
// return error
func (t *TargetPool) Switch(new bool) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Printf("start switch to %v\n", new)

	if t.useNew == new {
		return fmt.Errorf("already in %v", new)
	}

	// wait until all connection relased or destroyed
	timeout := time.After(15 * time.Second)

	for {
		stat := t.o.Stat()
		log.Printf("still in use: %d\n", stat.AcquiredConns())
		if stat.AcquiredConns() == 0 {
			break
		}

		select {
		case <-timeout:
			return errors.New("timeout")
		case <-time.After(1 * time.Second):
			continue
		}
	}

	//t.o.Close()

	// it should block until all old connection pool is done
	t.useNew = new
	return nil
}

// UseNew should block if we in the switching progress
func (t *TargetPool) UseNew() bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.useNew
}

// fake implementation of acquire
func (t *TargetPool) Acquire() (ctrl.Target, error) {
	c, err := t.o.Acquire(context.Background())
	if err != nil {
		return nil, err
	}

	return &Target{
		Conn:     c.Conn().PgConn().Conn(),
		pool:     t,
		poolConn: c,
	}, nil
}

// fake implementation of acquire new
func (t *TargetPool) AcquireNew() (ctrl.Target, error) {
	c, err := t.n.Acquire(context.Background())
	if err != nil {
		return nil, err
	}

	return &Target{
		Conn:     c.Conn().PgConn().Conn(),
		pool:     t,
		poolConn: c,
	}, nil
}

// fake implementation of release
func (t *TargetPool) Release(target *Target) {
	target.poolConn.Release()
}

func (t *TargetPool) Close() error {
	t.o.Close()
	t.n.Close()
	return nil
}

func (t *TargetPool) HijackOne() (*pgconn.HijackedConn, func(), error) {
	c, err := t.o.Acquire(context.Background())
	if err != nil {
		return nil, func() {}, err

	}

	h, err := c.Conn().PgConn().Hijack()

	return h, c.Release, err
}
