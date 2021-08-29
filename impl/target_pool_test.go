package impl

import "testing"

func TestTargetPoolNew(t *testing.T) {
	tp, err := NewTargetPool(
		"postgres://nakama:asdf@127.0.0.1:9000/nakama?sslmode=disable&pool_max_conns=10&pool_min_conns=3&pool_max_conn_idle_time=3s&pool_max_conn_lifetime=60s",
		"postgres://nakama:asdf@127.0.0.1:9000/nakama?sslmode=disable&pool_max_conns=10&pool_min_conns=3&pool_max_conn_idle_time=3s&pool_max_conn_lifetime=60s",
		false,
	)

	if err != nil {
		t.Error(err)
	}

	target, err := tp.Acquire()
	t.Error(target)
	t.Error(err)

}
