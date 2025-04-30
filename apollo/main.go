package apollo

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	//"errors"

	b64 "encoding/base64"

	"github.com/go-resty/resty/v2"
	"github.com/pelletier/go-toml/v2"
	"github.com/valyala/fastjson"
)

var (
	hc          *resty.Client
	dialTimeout = 5 * time.Second
)

type Apollo_t struct {
	Hc        *resty.Client
	Endpoints []string
	Name      string
	Pass      string
}

func New(endpoints []string, name, pass string) *Apollo_t {

	recs := new(Apollo_t)
	recs.Name = name
	recs.Pass = pass
	recs.Endpoints = endpoints

	transport := &http.Transport{
		MaxIdleConns:        6,           // Maximum number of idle connections
		MaxIdleConnsPerHost: 3,           // Maximum idle connections per host
		IdleConnTimeout:     dialTimeout, // Idle connection timeout
	}

	recs.Hc = resty.New().SetTransport(transport)
	recs.Hc.SetTimeout(dialTimeout)

	return recs
}

func (that *Apollo_t) GetToken(uri string) (string, error) {

	if that.Name == "" && that.Pass == "" {
		return "", nil
	}

	body := fmt.Sprintf(`{"name": "%s", "password": "%s"}`, that.Name, that.Pass)
	resp, err := that.Hc.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("http://" + uri + "/v3/auth/authenticate")

	if err != nil {
		fmt.Println("getToken err = ", err)
		return "", err
	}

	dd, err := fastjson.ParseBytes(resp.Body())
	if err != nil {
		fmt.Println("getToken fastjson.ParseBytes err = ", err)
		return "", err
	}
	if !dd.Exists("token") {
		return "", errors.New("etcd token error")
	}

	return string(dd.GetStringBytes("token")), nil
}

func (that *Apollo_t) GetContent(uri, key, token string) ([]byte, error) {

	sEnc := b64.StdEncoding.EncodeToString([]byte(key))

	resp, err := that.Hc.R().
		SetHeader("Authorization", token).
		SetHeader("Content-Type", "application/json").
		SetBody(`{"key": "` + sEnc + `"}`).
		Post("http://" + uri + "/v3/kv/range")

	if err != nil {
		fmt.Println("err = ", err)
		return nil, err
	}
	dd, err := fastjson.ParseBytes(resp.Body())
	if err != nil {
		fmt.Println("GetContent fastjson.ParseBytes err = ", err)
		return nil, err
	}
	if !dd.Exists("kvs") {
		return nil, errors.New("etcd key error")
	}

	str := string(dd.GetStringBytes("kvs", "0", "value"))

	sDec, err := b64.StdEncoding.DecodeString(str)
	return sDec, err
}

func (that *Apollo_t) Parse(key string, val interface{}) error {

	for _, v := range that.Endpoints {
		token, err := that.GetToken(v)
		if err != nil {
			fmt.Println("getToken err = ", err)
			continue
		}

		content, err := that.GetContent(v, key, token)
		if err != nil {
			fmt.Println("GetContent err = ", err)
			continue
		}
		return toml.Unmarshal(content, val)
	}

	return errors.New("没有可用的etcd服务器")
}
