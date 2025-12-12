package helper

import (
	"fmt"
	"log"

	"github.com/sony/sonyflake/v2"
)

var sf *sonyflake.Sonyflake

func init() {
	var err error
	st := sonyflake.Settings{}

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
