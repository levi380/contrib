package beanstalkpool

import (
	"context"
	"log"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/jackc/puddle/v2"
)

var pool *puddle.Pool[*beanstalk.Conn]

func New(addr string, maxIdle, maxActive int, dialTimeout time.Duration) {

	var err error

	constructor := func(context.Context) (*beanstalk.Conn, error) {
		return beanstalk.Dial("tcp", addr)
	}
	destructor := func(value *beanstalk.Conn) {
		value.Close()
	}
	maxPoolSize := int32(10)

	pool, err = puddle.NewPool(&puddle.Config[*beanstalk.Conn]{Constructor: constructor, Destructor: destructor, MaxSize: maxPoolSize})
	if err != nil {
		log.Fatal(err)
	}
}

func Get() (*puddle.Resource[*beanstalk.Conn], error) {

	return pool.Acquire(context.Background())
}

func Close() {
	pool.Close()
}
