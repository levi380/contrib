package validator

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"unsafe"

	"github.com/fxamacker/cbor/v2"
	"github.com/modern-go/reflect2"
	"github.com/valyala/fasthttp"
)

type spec func(str string) bool

var (
	structCache sync.Map // 用于缓存结构体字段信息
	specFunc    = map[string]spec{
		"url":   CtypeUrl,   // 验证网址格式
		"mail":  CtypeMail,  // 验证邮件格式
		"punct": CtypePunct, // 验证是否有标点符号

		"bankcard": PhBankCard, // 验证是否有标点符号
		"phone":    PhPhone,    // 验证是否有标点符号

		//	"alpha":      ctypeAlpha,      // 验证是否纯字母
		"digit":      CtypeDigit,      // 验证是否纯数字
		"digitComma": CtypeDigitComma, // 验证是否数字加逗号
		"alnum":      CtypeAlnum,      // 验证是否数字加英文
		"amount":     CtypeAmount,     //  验证是否数字英文-_
		"date":       CtypeDate,       // 验证是否时间格式"2022-02-12"
		"datetime":   CtypeDateTime,   // 验证是否时间格式"2022-02-12 00:00:00"
	}
)

type FieldInfo struct {
	Name     string
	Rule     string
	Field    []string
	Required bool
	Type     reflect2.Type
	Index    []int
}

func getStructFields(typ reflect2.Type) []FieldInfo {
	// 检查缓存
	if cached, ok := structCache.Load(typ); ok {
		return cached.([]FieldInfo)
	}

	// 如果缓存中没有，则解析结构体
	structType := typ.(reflect2.StructType)
	numField := structType.NumField()
	fields := make([]FieldInfo, numField)

	for i := 0; i < numField; i++ {
		field := structType.Field(i)
		fields[i] = FieldInfo{
			Rule: field.Tag().Get("rule"),
			Name: field.Tag().Get("json"),
			//Name:     strings.ToLower(field.Name()),
			Required: field.Tag().Get("required") == "true",
			Type:     field.Type(),
			Index:    []int{i}, // 存储字段的索引
		}

		if fields[i].Rule == "enum" {
			fields[i].Field = strings.Split(field.Tag().Get("field"), ",")
		}
	}

	// 存入缓存
	structCache.Store(typ, fields)

	return fields
}

func BindStruct(payload []byte, out interface{}) error {
	err := cbor.Unmarshal(payload, out)
	if err != nil {
		return err
	}

	outPtr := reflect2.PtrOf(out)
	if outPtr == nil {
		return fmt.Errorf("out must be a pointer to a struct")
	}

	userType := reflect2.TypeOf(out).(*reflect2.UnsafePtrType).Elem()
	structType := userType.(reflect2.StructType)
	fields := getStructFields(userType)
	for _, field := range fields {

		fieldType := userType.(reflect2.StructType).FieldByIndex(field.Index)

		if field.Required && field.Type.Kind() == reflect.String {
			v := fieldType.Get(structType.PackEFace(outPtr))
			if cb, ok := specFunc[field.Rule]; ok {

				if value, ok := v.(*string); ok {
					success := cb(*value)
					if !success {
						return errors.New(field.Name)
					}
				}
			} else {
				if value, ok := v.(*string); ok {
					str := strings.TrimSpace(*value)
					if len(str) == 0 {
						return errors.New(field.Name)
					}
				}
			}
		}
	}

	return nil
}

func Bind(args *fasthttp.Args, out interface{}) error {
	outPtr := reflect2.PtrOf(out)
	if outPtr == nil {
		return fmt.Errorf("out must be a pointer to a struct")
	}

	userType := reflect2.TypeOf(out).(*reflect2.UnsafePtrType).Elem()

	fields := getStructFields(userType)

	for _, field := range fields {

		if field.Required {
			value := string(args.Peek(field.Name))
			if field.Rule == "enum" {
				if !slices.Contains(field.Field, value) {
					return errors.New(field.Name)
				}

			} else {
				if cb, ok := specFunc[field.Rule]; ok {
					success := cb(value)
					if !success {
						return errors.New(field.Name)
					}
				}
			}
		}

		if args.Has(field.Name) {

			if !field.Required && field.Rule != "" {
				//在非必填的情况下，验证
				value := string(args.Peek(field.Name))
				if field.Rule == "enum" {
					if !slices.Contains(field.Field, value) {
						return errors.New(field.Name)
					}

				} else {
					if cb, ok := specFunc[field.Rule]; ok {
						success := cb(value)
						if !success {
							return errors.New(field.Name)
						}
					}
				}
			}

			fieldType := userType.(reflect2.StructType).FieldByIndex(field.Index)

			switch field.Type.Kind() {
			case reflect.Bool:
				val := args.GetBool(field.Name)
				fieldType.UnsafeSet(unsafe.Pointer(outPtr), reflect2.PtrOf(val))
			case reflect.Int64, reflect.Int, reflect.Int32:
				if val, err := args.GetUint(field.Name); err == nil {
					fieldType.UnsafeSet(unsafe.Pointer(outPtr), reflect2.PtrOf(val))
				}
			case reflect.Float32, reflect.Float64:
				if val, err := args.GetUfloat(field.Name); err == nil {
					fieldType.UnsafeSet(unsafe.Pointer(outPtr), reflect2.PtrOf(val))
				}
			case reflect.String:
				val := string(args.Peek(field.Name))
				fieldType.UnsafeSet(unsafe.Pointer(outPtr), reflect2.PtrOf(val))

			}
		}
	}

	return nil
}
