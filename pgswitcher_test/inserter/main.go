package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

var point uint64 = 3

func main() {
	pool, err := pgxpool.Connect(context.Background(), "postgres://system:123456@127.0.0.1:5440/trial01?sslmode=disable&pool_max_conns=4")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		go insert(pool)
	}

	for {
	}
}

func insert(p *pgxpool.Pool) {
	for {
		name := atomic.AddUint64(&point, 1)
		_, err := p.Exec(context.Background(), fmt.Sprintf("INSERT INTO table01(name) VALUES(%d)", name))
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
