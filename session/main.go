package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
	"lukechampine.com/frand"
)

var (
	prefix  = "t:"
	rr      *frand.RNG
	ctx     = context.Background()
	client  *redis.Client
	expires [2]time.Duration
)

func New(reddb *redis.Client) {

	rr = frand.New()
	client = reddb
	expires[0] = time.Duration(30) * time.Minute
	expires[1] = time.Duration(30) * time.Minute
}

func Set(ty int, name string, seed uint32) (string, error) {

	uuid := fmt.Sprintf("TI:%s", name)

	key := fmt.Sprintf("%x", rr.Entropy128())
	val, err := client.Get(ctx, uuid).Result()

	key = prefix + key

	//ex := fmt.Sprintf("%d", ts.Unix())
	pipe := client.Pipeline()

	if err != redis.Nil && len(val) > 0 {
		//同一个用户，一个时间段，只能登录一个
		pipe.Del(ctx, val)
	}

	pipe.Set(ctx, uuid, key, 720*time.Hour)
	pipe.Set(ctx, key, name, expires[ty])
	//pipe.ZAdd(ctx, "online", vv)
	_, err = pipe.Exec(ctx)

	return key, err
}

func Offline(uids string) error {

	uuid := fmt.Sprintf("TI:%s", uids)
	sid, err := client.Get(ctx, uuid).Result()
	if err != nil {
		return err
	}

	err = client.Unlink(ctx, sid, uuid).Err()
	return err
}

func Get(fctx *fasthttp.RequestCtx) ([]byte, error) {

	key := string(fctx.Request.Header.Peek("t"))
	if len(key) == 0 {

		//fmt.Println("string(fctx.Request.Header.Peek(t)) = ", string(fctx.Request.Header.Peek("t")))
		key = string(fctx.QueryArgs().Peek("t"))
		if len(key) == 0 {
			//fmt.Println("string(fctx.QueryArgs().Peek(t)) = ", string(fctx.QueryArgs().Peek("t")))
			return nil, errors.New("does not exist")
		}
	}

	return client.Get(ctx, key).Bytes()

}

func ExpireAt(fctx *fasthttp.RequestCtx, ty int) error {

	key := string(fctx.Request.Header.Peek("t"))
	if len(key) == 0 {

		key = string(fctx.QueryArgs().Peek("t"))
		if len(key) == 0 {
			return errors.New("does not exist")
		}
	}

	return client.ExpireXX(ctx, key, expires[ty]).Err()
}

func AdminSet(value []byte, uid string) (string, error) {

	uuid := fmt.Sprintf("TI:%s", uid)
	key := fmt.Sprintf("%x", rr.Entropy128())

	val, err := client.Get(ctx, uuid).Result()

	pipe := client.Pipeline()

	if err != redis.Nil {
		//同一个用户，一个时间段，只能登录一个
		pipe.Unlink(ctx, val)
	}
	pipe.Set(ctx, uuid, key, time.Duration(100)*time.Hour)
	pipe.Set(ctx, key, value, time.Duration(4320)*time.Minute)
	_, err = pipe.Exec(ctx)

	return key, err
}

func AdminGet(fctx *fasthttp.RequestCtx) ([]byte, error) {

	key := string(fctx.Request.Header.Peek("t"))
	if len(key) == 0 {
		key = string(fctx.QueryArgs().Peek("t"))
		if len(key) == 0 {
			return nil, errors.New("does not exist")
		}
	}

	return client.Get(ctx, key).Bytes()
}
