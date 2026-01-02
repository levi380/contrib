package helper

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/valyala/fasthttp"
	"lukechampine.com/frand"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Response struct {
	Status bool `json:"status" cbor:"status"`
	Data   any  `json:"data" cbor:"data"`
}

func Serialize(state bool, data any) ([]byte, error) {

	var (
		b   []byte
		err error
	)

	res := Response{
		Status: state,
		Data:   data,
	}
	b, err = cbor.Marshal(res)

	return b, err
}

func Echo(fctx *fasthttp.RequestCtx, status bool, data any) {

	var (
		b   []byte
		err error
	)
	fctx.SetContentType("application/x-protobuf")
	fctx.SetStatusCode(200)
	fctx.Response.Header.Set("Pragma", "no-cache")
	fctx.Response.Header.Set("Expires", "0")
	fctx.Response.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	res := Response{
		Status: status,
		Data:   data,
	}

	b, err = cbor.Marshal(res)

	if err != nil {
		fctx.SetBodyString(err.Error())
		return
	}

	fctx.SetBody(b)
}

/*
func GenId() string {

	var min uint64 = 0
	var max uint64 = 9

	return fmt.Sprintf("%d%d", Cputicks(), frand.Uint64n(max-min)+min)
}

func GenLongId() string {

	var min uint64 = 100000
	var max uint64 = 999999

	id := fmt.Sprintf("%d%d", Cputicks(), frand.Uint64n(max-min)+min)
	return id[0:18]
}
*/

func RandomKey(length int) string {

	id := make([]byte, length)
	for i := range id {
		//id[i] = alphabet[frand.Intn(len(alphabet))]
		id[i] = alphabet[frand.Intn(62)]
	}
	return string(id)
}


