package validator

import (
	//"fmt"
	"net"
	"net/mail"
	"net/url"
	"time"
	"unicode"

	"github.com/valyala/fastjson"
)

// IsValidPhilippinesBankCard 检查是否为有效的菲律宾银行卡号码
func PhBankCard(str string) bool {
	if len(str) != 16 {
		return false
	}

	return CtypeDigit(str)
}

func PhPhone(str string) bool {

	if len(str) != 10 {
		return false
	}

	if str[:1] != "9" {
		return false
	}

	return CtypeDigit(str)
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isAlpha(r rune) bool {
	if r >= 'A' && r <= 'Z' {
		return true
	} else if r >= 'a' && r <= 'z' {
		return true
	}
	return false
}

func CtypeJson(str string) bool {

	err := fastjson.Validate(str)
	return err == nil
}

func CtypeIp(str string) bool {
	// net.ParseIP 尝试解析 IP 字符串
	// 如果解析成功，则返回一个非 nil 的 net.IP 类型值
	// 如果解析失败，则返回 nil
	ip := net.ParseIP(str)

	// 如果 ip 不为 nil，则表示解析成功，是一个有效的 IP 地址
	return ip != nil
}

func CtypeUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func CtypeMail(str string) bool {
	_, err := mail.ParseAddress(str)

	return err == nil
}

func CtypeDate(str string) bool {
	_, err := time.Parse("2006-01-02", str)

	return err == nil
}

func CtypeDateTime(str string) bool {
	_, err := time.Parse("2006-01-02 15:04:05", str)

	return err == nil
}

func CtypeDigitComma(str string) bool {
	if str == "" {
		return false
	}
	for _, r := range str {
		if !isDigit(r) && r != ',' {
			return false
		}
	}
	return true
}

func CtypeAmount(str string) bool {
	if str == "" {
		return false
	}
	for _, r := range str {
		if !isDigit(r) && r != '.' {
			return false
		}
	}
	return true
}

func CtypeDigit(str string) bool {
	if str == "" {
		return false
	}
	for _, r := range str {
		if !isDigit(r) {
			return false
		}
	}
	return true
}

func CtypeAlphaComma(str string) bool {
	if str == "" {
		return false
	}

	for _, r := range str {
		if !isAlpha(r) && r != ',' {
			return false
		}
	}

	return true
}

func CtypePunct(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsPunct(r) || unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func CtypeAlnum(str string) bool {
	if str == "" {
		return false
	}
	for _, r := range str {
		if !isDigit(r) && !isAlpha(r) {
			return false
		}
	}

	return true
}
