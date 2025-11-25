package helper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sony/sonyflake/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var sf *sonyflake.Sonyflake

func Initial(endpoints []string) {

	var (
		mctx  = context.Background()
		mutex *concurrency.Mutex
	)
	// 根据etcd 生成work id
	client, err := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: time.Second * 3})
	if err != nil {
		log.Fatalf("客户端初始化失败:%v\n", err)
	}

	sess, err := concurrency.NewSession(client, concurrency.WithTTL(30))
	if err != nil {
		log.Fatalf("Session初始化失败:%v\n", err)
		return
	}

	workId := 0
	for i := 1; i < 4095; i++ {
		// 获取指定前缀的锁对象
		key := fmt.Sprintf("my-lock/%d", i)
		mutex = concurrency.NewMutex(sess, key)

		// 加锁默认等待3s
		err = mutex.TryLock(mctx)
		if err == nil {
			workId = i
			fmt.Printf("尝试获取work id:%s\n", key)
			break
		}
	}


	st := sonyflake.Settings{}
	st.MachineID = workId
	sf, err = sonyflake.New(st) // 使用默认设置
	if err != nil {
		log.Fatalf("failed to create sonyflake: %v", err)
	}
}

func GenId() string {

	id, err := sf.NextID()
	if err != nil {
		fmt.Println("GenId NextID err = ", err)
		return fmt.Sprintf("%d", Cputicks())
	}

	return fmt.Sprintf("%d", id)
}
