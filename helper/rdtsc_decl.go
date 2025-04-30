package helper

import (
	"crypto/tls"
	"github.com/valyala/fasthttp"
	"time"
)

var (
	fc = &fasthttp.Client{
		MaxConnsPerHost: 60000,
		TLSConfig:       &tls.Config{InsecureSkipVerify: true},
	}
)

func HttpDoTimeout(requestBody []byte, method string, requestURI string, headers map[string]string, timeout time.Duration) ([]byte, int, error) {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(requestURI)
	req.Header.SetMethod(method)

	switch method {
	case "POST", "PUT", "DELETE":
		req.SetBody(requestBody)
	}

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// time.Second * 30
	err := fc.DoTimeout(req, resp, timeout)

	return resp.Body(), resp.StatusCode(), err
}
