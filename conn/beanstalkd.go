package conn


import (
	"log"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/silenceper/pool"
)

type Connection struct {
	Conn pool.Pool
}

// ryqueue
func BeanNew(addr string) *Connection {

	factory := func() (interface{}, error) { return beanstalk.Dial("tcp", addr) }
	close := func(v interface{}) error { return v.(*beanstalk.Conn).Close() }

	//ping Specify the method to detect whether the connection is invalid
	//ping := func(v interface{}) error { return nil }

	// Create a connection pool: Initialize the number of connections to 5, the maximum idle connection is 20, and the maximum concurrent connection is 30
	poolConfig := &pool.Config{
		InitialCap: 3,
		MaxIdle:    10,
		MaxCap:     15,
		Factory:    factory,
		Close:      close,
		//Ping:       ping,
		//The maximum idle time of the connection, the connection exceeding this time will be closed, which can avoid the problem of automatic failure when connecting to EOF when idle
		IdleTimeout: 15 * time.Second,
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		log.Fatalf("ryqueue NewChannelPool err = %s", err.Error())

	}
	return &Connection{
		Conn: p,
	}
}

func (c *Connection) Put(name string, body []byte, pri uint32, delay, ttr time.Duration) (id uint64, err error) {

	v, err := c.Conn.Get()
	if err != nil {
		return 0, err
	}
	defer c.Conn.Put(v)
	tube := beanstalk.NewTube(v.(*beanstalk.Conn), name)

	id, err1 := tube.Put(body, pri, delay, ttr)
	if err1 != nil {
		return 0, err1
	}
	return id, nil
}

func (c *Connection) Reserve(name []string, timeout time.Duration) (id uint64, body []byte, err error) {

	v, err := c.Conn.Get()
	if err != nil {
		return 0, nil, err
	}
	defer c.Conn.Put(v)
	tube := beanstalk.NewTubeSet(v.(*beanstalk.Conn), name...)

	id, body, err1 := tube.Reserve(timeout)
	if err1 != nil {
		return 0, nil, err1
	}
	return id, body, err1
}

func (c *Connection) Delete(id uint64) error {

	v, err := c.Conn.Get()
	if err != nil {
		return err
	}
	defer c.Conn.Put(v)
	err1 := v.(*beanstalk.Conn).Delete(id)
	return err1
}

func (c *Connection) Release(id uint64, pri uint32, delay time.Duration) error {

	v, err := c.Conn.Get()
	if err != nil {
		return err
	}
	defer c.Conn.Put(v)
	err1 := v.(*beanstalk.Conn).Release(id, pri, delay)
	return err1
}

func (c *Connection) Bury(id uint64, pri uint32) error {

	v, err := c.Conn.Get()
	if err != nil {
		return err
	}
	defer c.Conn.Put(v)
	err1 := v.(*beanstalk.Conn).Bury(id, pri)
	return err1
}
