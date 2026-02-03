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
	expires [3]time.Duration
)

func New(reddb *redis.Client) {

	rr = frand.New()
	client = reddb
	expires[0] = time.Duration(30) * time.Minute
	expires[1] = time.Duration(30) * time.Minute
	expires[2] = time.Duration(72) * time.Hour
}

func Set(fctx *fasthttp.RequestCtx, loc *time.Location, ty int, name string, seed uint32) (string, error) {

	device := string(fctx.Request.Header.Peek("d"))

	uuid := fmt.Sprintf("T:%s:%s", name, device)

	key := fmt.Sprintf("%x", rr.Entropy128())
	val, err := client.Get(ctx, uuid).Result()

	key = prefix + key

	day := fctx.Time().In(loc).Format("01-02")
	ckey := fmt.Sprintf("%s-member-login", day)

	//ex := fmt.Sprintf("%d", ts.Unix())
	pipe := client.Pipeline()

	if err != redis.Nil && len(val) > 0 {
		//同一个用户，一个时间段，只能登录一个
		pipe.Del(ctx, val)
	}

	pipe.PFAdd(ctx, ckey, name)
	pipe.ExpireXX(ctx, ckey, time.Duration(48)*time.Hour)
	pipe.Set(ctx, uuid, key, 720*time.Hour)
	pipe.Set(ctx, key, name, expires[ty])
	/*
		if ty == 0 {
			pipe.HSet(ctx, "online_member", name, "1")
			pipe.HExpire(ctx, "online_member", expires[ty], name)
		}
	*/
	//pipe.ZAdd(ctx, "online", vv)
	_, err = pipe.Exec(ctx)

	return key, err
}

func Offline(uid string) error {

	keys := []string{}

	uuid := fmt.Sprintf("T:%s", uid)

	pipe1 := client.Pipeline()
	uuid24_temp := pipe1.Get(ctx, uuid+":24")
	uuid25_temp := pipe1.Get(ctx, uuid+":25")
	uuid26_temp := pipe1.Get(ctx, uuid+":26")
	uuid27_temp := pipe1.Get(ctx, uuid+":27")
	uuid28_temp := pipe1.Get(ctx, uuid+":28")

	pipe1.Exec(ctx)
	uuid24, err24 := uuid24_temp.Result()
	if err24 == nil {
		keys = append(keys, uuid24)
	}
	uuid25, err25 := uuid25_temp.Result()
	if err25 == nil {
		keys = append(keys, uuid25)
	}
	uuid26, err26 := uuid26_temp.Result()
	if err26 == nil {
		keys = append(keys, uuid26)
	}
	uuid27, err27 := uuid27_temp.Result()
	if err27 == nil {
		keys = append(keys, uuid27)
	}
	uuid28, err28 := uuid28_temp.Result()
	if err28 == nil {
		keys = append(keys, uuid28)
	}

	pipe2 := client.Pipeline()
	for _, key := range keys {
		pipe2.Unlink(ctx, key)
	}
	//pipe.HDel(ctx, "online_member", sid)
	pipe2.Del(ctx, "onlines:"+uid)
	pipe2.Exec(ctx)
	return nil
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

	uid, err := client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	pipe := client.Pipeline()
	pipe.ExpireXX(ctx, key, expires[ty])
	//pipe.HExpire(ctx, "online_member", expires[ty], uid)
	pipe.ExpireXX(ctx, "onlines:"+uid, expires[ty])
	pipe.Exec(ctx)

	return nil
}

func Del(fctx *fasthttp.RequestCtx) error {

	key := string(fctx.Request.Header.Peek("t"))
	if len(key) == 0 {

		key = string(fctx.QueryArgs().Peek("t"))
		if len(key) == 0 {
			return errors.New("does not exist")
		}
	}

	uid, err := client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	pipe := client.Pipeline()
	pipe.Unlink(ctx, key)
	pipe.Unlink(ctx, "onlines:"+uid)
	pipe.Exec(ctx)

	return nil
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
	pipe.Set(ctx, uuid, key, expires[1])
	pipe.Set(ctx, key, value, expires[1])
	_, err = pipe.Exec(ctx)

	return key, err
}

/*
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
*/
