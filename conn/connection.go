package conn

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	xxteaKey = ""
	ctx      = context.Background()
)

func Use(xxKey string) {
	xxteaKey = xxKey
}

func Chacha20Encode(msg, pass string) string {

	key := sha256.Sum256([]byte(pass))
	aead, _ := chacha20poly1305.NewX(key[:])
	nonce := make([]byte, chacha20poly1305.NonceSizeX)

	return base64.StdEncoding.EncodeToString(aead.Seal(nil, nonce, []byte(msg), nil))
}

func Chacha20Decode(ciphertext, pass string) ([]byte, error) {

	decode, _ := base64.StdEncoding.DecodeString(ciphertext)
	key := sha256.Sum256([]byte(pass))
	aead, _ := chacha20poly1305.NewX(key[:])
	nonce := make([]byte, chacha20poly1305.NonceSizeX)

	return aead.Open(nil, nonce, decode, nil)
}

func InitDB(dsn string, maxIdleConn, maxOpenConn int) *sql.DB {

	if xxteaKey != "" {
		dst, err := Chacha20Decode(dsn, xxteaKey)
		if err != nil {
			fmt.Println("InitDB", dsn, err)
			log.Fatalln(err)
		}
		dsn = string(dst)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	db.SetMaxOpenConns(maxOpenConn)
	db.SetMaxIdleConns(maxIdleConn)
	db.SetConnMaxLifetime(time.Second * 30)
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("mysql db.Ping err = ", err)
	return db
}

func InitDBX(dsn string, maxIdleConn, maxOpenConn int) *sqlx.DB {

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	db.SetMaxOpenConns(maxOpenConn)
	db.SetMaxIdleConns(maxIdleConn)
	db.SetConnMaxLifetime(time.Second * 30)
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("mysql db.Ping err = ", err)
	return db
}

func InitRedisSentinel(dsn []string, username, psd, name string, db, poolSize int) *redis.Client {

	cfg := &redis.FailoverOptions{
		MasterName:    name,
		SentinelAddrs: dsn,
		Password:      psd, // no password set
		DB:            db,  // use default DB
		DialTimeout:   10 * time.Second,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		PoolSize:      poolSize,
		PoolTimeout:   30 * time.Second,
		MaxRetries:    2,
		MaxIdleConns:  20,
		//MaxActiveConns:        20,
		ConnMaxLifetime:       5 * time.Minute,
		ContextTimeoutEnabled: false,
		UnstableResp3:         true,
	}

	if username != "" {
		cfg.Username = username
	}
	reddb := redis.NewFailoverClient(cfg)
	pong, err := reddb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("InitRedisSentinel failed: %s", err.Error())
	}
	fmt.Println(pong, err)

	return reddb
}

func InitRedis(dsn string, passwd string, db,poolSize int) *redis.Client {

	reddb := redis.NewClient(&redis.Options{
		Addr: dsn,
		//Username: "user",
		DB : db,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     poolSize,
		PoolTimeout:  30 * time.Second,
		MaxRetries:   2,
		MaxIdleConns: 20,
		//MaxActiveConns:        20,
		ConnMaxLifetime:       5 * time.Minute,
		ContextTimeoutEnabled: false,
		Password:              passwd,
	})
	pong, err := reddb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("InitRedisSentinel failed: %s", err.Error())
	}
	fmt.Println(pong, err)

	return reddb
}
