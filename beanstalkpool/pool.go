package beanstalkpool

import (
	"context"
	"log"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/jackc/puddle/v2"
)

var (
	pool *puddle.Pool[*beanstalk.Conn]
	ctx  = context.Background()
)

func New(addr string, maxIdle, maxActive int, dialTimeout time.Duration) {

	var err error

	constructor := func(context.Context) (*beanstalk.Conn, error) {
		return beanstalk.Dial("tcp", addr)
	}
	destructor := func(value *beanstalk.Conn) {
		value.Close()
	}
	maxPoolSize := int32(maxIdle)

	pool, err = puddle.NewPool(&puddle.Config[*beanstalk.Conn]{Constructor: constructor, Destructor: destructor, MaxSize: maxPoolSize})
	if err != nil {
		log.Fatal(err)
	}
}

func Put(topic string, body []byte, pri uint32, delay, ttr time.Duration) (id uint64, err error) {

	res, err := pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer res.Release()

	tube := beanstalk.NewTube(res.Value(), topic)

	id, err1 := tube.Put(body, pri, delay, ttr)
	if err1 != nil {
		return 0, err1
	}
	return id, nil
}

func Reserve(topic []string, timeout time.Duration) (id uint64, body []byte, err error) {

	res, err := pool.Acquire(ctx)
	if err != nil {
		return 0, nil, err
	}
	defer res.Release()

	tube := beanstalk.NewTubeSet(res.Value(), topic...)
	return tube.Reserve(timeout)
}

func Delete(id uint64) error {

	res, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer res.Release()

	return res.Value().Delete(id)
}

func Release(id uint64, pri uint32, delay time.Duration) error {

	res, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer res.Release()

	return res.Value().Release(id, pri, delay)
}

func Bury(id uint64, pri uint32) error {

	res, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer res.Release()
	return res.Value().Bury(id, pri)
}

/*
func Get() (*puddle.Resource[*beanstalk.Conn], error) {

	return pool.Acquire(ctx)
}
*/

func Close() {
	pool.Close()
}
