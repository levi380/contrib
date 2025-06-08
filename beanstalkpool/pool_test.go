package beanstalkpool

import (
	"context"
	"testing"
	"time"
)

func TestPoolBasic(t *testing.T) {
	pool := New("127.0.0.1:11300", 2, 5, time.Second)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	pool.Put(conn)
}