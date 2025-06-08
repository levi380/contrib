package beanstalkpool

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/beanstalkd/go-beanstalk"
)

type Pool struct {
	addr        string
	maxIdle     int
	maxActive   int
	dialTimeout time.Duration

	mu     sync.Mutex
	idle   chan *beanstalk.Conn
	active int
	closed bool
}

func New(addr string, maxIdle, maxActive int, dialTimeout time.Duration) *Pool {
	return &Pool{
		addr:        addr,
		maxIdle:     maxIdle,
		maxActive:   maxActive,
		dialTimeout: dialTimeout,
		idle:        make(chan *beanstalk.Conn, maxIdle),
	}
}

func (p *Pool) Get(ctx context.Context) (*beanstalk.Conn, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("beanstalk pool is closed")
	}
	select {
	case conn := <-p.idle:
		p.mu.Unlock()
		if isClosed(conn) {
			return p.dial()
		}
		return conn, nil
	default:
		if p.active >= p.maxActive {
			p.mu.Unlock()
			select {
			case conn := <-p.idle:
				if isClosed(conn) {
					return p.dial()
				}
				return conn, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		p.active++
		p.mu.Unlock()
		return p.dial()
	}
}

func (p *Pool) Put(conn *beanstalk.Conn) {
	if conn == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		conn.Close()
		p.active--
		return
	}

	select {
	case p.idle <- conn:
	default:
		conn.Close()
		p.active--
	}
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true
	close(p.idle)
	for conn := range p.idle {
		conn.Close()
	}
	p.active = 0
}

func (p *Pool) dial() (*beanstalk.Conn, error) {
	conn, err := net.DialTimeout("tcp", p.addr, p.dialTimeout)
	if err != nil {
		p.mu.Lock()
		p.active--
		p.mu.Unlock()
		return nil, err
	}
	return beanstalk.NewConn(conn), nil
}

func isClosed(c *beanstalk.Conn) bool {
	_, err := c.Stats()
	return err != nil
}