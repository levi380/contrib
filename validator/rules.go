package validator

import (
	//"fmt"
	"net/mail"
	"net/url"
	"time"
	"unicode"
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
