### 使用示例：

待测函数：

```golang
package main

import (
	"fmt"
	"log"

	"validator"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

//Fd = formData
type indexFd struct {
	Name string `required:"false" rule:"alnum"`
	Age  int    `required:"true" rule:"digit"`
}

func Index(ctx *fasthttp.RequestCtx) {
	var q queryFd
	err := validator.Bind(ctx.QueryArgs(), &q)
	if err != nil {
		// 处理错误
		fmt.Fprintf(ctx, "Error: %v\n", err)
		return
	}
	fmt.Printf("%+v\n", q)
	ctx.WriteString("Welcome!")
}


func main() {
	r := router.New()
	r.GET("/", Index)

	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}

```

### rule规则说明

|字段|说明|
|-|-|
|url|URL格式|
|mail|电子邮件格式|
|alpha|英文字母|
|digit|数字|
|datetime|2022-02-01 13:02:30|
|alnumDash|数字、字母、-、_|
|alnum|数字和英文字母|


### in

rule=in&value=1,2,3,4

### 长度限制

|字段|说明|
|min=2|最小长度|
|max=10|最大长度|
|range=1,3|范围|

