package conn

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/qiniu/qmgo"
	qnOpts "github.com/qiniu/qmgo/options"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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

func InitMongo(ctx context.Context, url, username, passwd, dbname string) (*qmgo.Client, *qmgo.Database) {

	var (
		timeout     int64  = 2000
		maxPoolSize uint64 = 100
		minPoolSize uint64 = 0
		durl               = []byte(url)
		dusername          = []byte(username)
		dpasswd            = []byte(passwd)
		err         error
	)
	if xxteaKey != "" {
		durl, err = Chacha20Decode(url, xxteaKey)
		if err != nil {
			fmt.Println("InitMongo", url, err)
			log.Fatalln(err)
		}

		dusername, err = Chacha20Decode(username, xxteaKey)
		if err != nil {
			fmt.Println("InitMongo", username, err)
			log.Fatalln(err)
		}

		dpasswd, err = Chacha20Decode(passwd, xxteaKey)
		if err != nil {
			fmt.Println("InitMongo", passwd, err)
			log.Fatalln(err)
		}
	}

	clientOptions := &options.ClientOptions{}
	// 设置认证信息
	credential := options.Credential{
		Username:   string(dusername),
		Password:   string(dpasswd),
		AuthSource: dbname,
	}
	clientOptions.SetAuth(credential)
	opts := qnOpts.ClientOptions{
		ClientOptions: clientOptions,
	}

	cfg := qmgo.Config{
		Uri:              string(durl),
		ConnectTimeoutMS: &timeout,
		MaxPoolSize:      &maxPoolSize,
		MinPoolSize:      &minPoolSize,
		ReadPreference:   &qmgo.ReadPref{Mode: readpref.SecondaryMode, MaxStalenessMS: 100 * 1000},
	}

	cli, err := qmgo.NewClient(ctx, &cfg, opts)
	if err != nil {
		log.Fatalf("initMongoClient failed: %s", err.Error())
	}
	err = cli.Ping(5)
	if err != nil {
		log.Fatalf("MongoClient ping failed: %s", err.Error())
	}

	return cli, cli.Database(dbname)
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

func InitRedisSentinelCluster(dsn []string, psd, name string, db, poolSize int) *redis.ClusterClient {

	var (
		dst = []byte(psd)
		err error
	)

	if xxteaKey != "" {
		dst, err = Chacha20Decode(psd, xxteaKey)
		if err != nil {
			fmt.Println("InitRedisSentinel", psd, err)
			log.Fatalln(err)
		}
	}

	reddb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:            name,
		SentinelAddrs:         dsn,
		Password:              string(dst), // no password set
		DB:                    db,          // use default DB
		Protocol:              2,
		ReplicaOnly:           true,
		DialTimeout:           10 * time.Second,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		PoolSize:              poolSize,
		PoolTimeout:           30 * time.Second,
		MaxRetries:            2,
		MaxIdleConns:          20,
		ConnMaxLifetime:       5 * time.Minute,
		ContextTimeoutEnabled: false,
	})
	pong, err := reddb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("InitRedisSentinel failed: %s", err.Error())
	}
	fmt.Println(pong, err)

	return reddb
}

func InitRedisSentinel(dsn []string, psd, name string, db, poolSize int) *redis.Client {

	var (
		dst = []byte(psd)
		err error
	)
	if xxteaKey != "" {
		dst, err = Chacha20Decode(psd, xxteaKey)
		if err != nil {
			fmt.Println("InitRedisSentinel", psd, err)
			log.Fatalln(err)
		}
	}

	reddb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    name,
		SentinelAddrs: dsn,
		Password:      string(dst), // no password set
		DB:            db,          // use default DB
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
	})
	pong, err := reddb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("InitRedisSentinel failed: %s", err.Error())
	}
	fmt.Println(pong, err)

	return reddb
}

func InitRedisCluster(dsn []string, passwd string, poolSize int) *redis.ClusterClient {

	reddb := redis.NewClusterClient(&redis.ClusterOptions{

		Addrs:        dsn,
		Password:     passwd, // no password set
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
	})
	pong, err := reddb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("InitRedisSentinel failed: %s", err.Error())
	}
	fmt.Println(pong, err)

	return reddb
}
