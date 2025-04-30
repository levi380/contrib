package helper

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/forgoer/openssl"
	//"github.com/google/uuid"
	//md5simd "github.com/minio/md5-simd"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
	"lukechampine.com/frand"

	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

/*
func UUID() string {
	return fmt.Sprintf("%v", uuid.New())
}

// md5
func MD5Hash(text string) string {

	server := md5simd.NewServer()
	md5Hash := server.NewHash()
	_, _ = md5Hash.Write([]byte(text))
	digest := md5Hash.Sum([]byte{})
	encrypted := hex.EncodeToString(digest)

	server.Close()
	md5Hash.Close()

	return encrypted
}
*/
// sha1
func Sha1Sum(s string) []byte {

	h := sha1.New()
	h.Write([]byte(s))

	return h.Sum(nil)
}

func HmacSha(source string, key string) string {

	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(source))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func HmacSha256(source string, key string) string {

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(source))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// sha256
func Sha256sum(param []byte) string {

	h := sha256.New()
	h.Write(param)

	return fmt.Sprintf("%x", h.Sum(nil))
}

func RsaEncrypt(privateKey, origData []byte) []byte {

	//设置私钥
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil
	}

	prkI, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil
	}

	priv := prkI.(*rsa.PrivateKey)
	encodeByte, _ := rsa.SignPKCS1v15(crand.Reader, priv, crypto.MD5, origData)

	return encodeByte
}

func rsaSha256Sign(privateKey, origData []byte) []byte {

	//设置私钥
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil
	}

	prkI, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil
	}

	h := sha256.New()
	h.Write(origData)
	data := h.Sum(nil)

	encodeByte, _ := rsa.SignPKCS1v15(crand.Reader, prkI, crypto.SHA256, data)

	return encodeByte
}

func rsaSha256Very(publicKey, data, signData []byte) bool {

	block, _ := pem.Decode(publicKey)
	if block == nil {
		return false
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signData)
	if err != nil {
		return false
	}
	return true
}

func rsaMd5Sign(privateKey, origData []byte) string {

	//设置私钥
	block, _ := pem.Decode(privateKey)
	if block == nil {
		fmt.Println("rsaMd5Sign block", block)
		return ""
	}

	prkI, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("rsaMd5Sign prkI error", prkI, err)
		return ""
	}

	h := md5.New()
	h.Write(origData)
	data := h.Sum(nil)

	encodeByte, _ := rsa.SignPKCS1v15(crand.Reader, prkI.(*rsa.PrivateKey), crypto.MD5, data)
	return base64.RawURLEncoding.EncodeToString(encodeByte)
}

func AesECBEncrypt(src, key []byte) (string, error) {
	dst, err := openssl.AesECBEncrypt(src, key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(dst), nil
}

func AesECBDecrypt(src, key []byte) ([]byte, error) {
	decode, _ := base64.StdEncoding.DecodeString(string(src))
	return openssl.AesECBDecrypt(decode, key, openssl.PKCS7_PADDING)
}

func JDBAesCBCEncrypt(src, key, iv []byte) (string, error) {
	dst, err := openssl.AesCBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(dst), nil
}

func JDBAesCBCDecrypt(src, key, iv []byte) ([]byte, error) {
	decode, err := base64.RawURLEncoding.DecodeString(string(src))
	if err != nil {
		return []byte{}, err
	}

	return openssl.AesCBCDecrypt(decode, key, iv, openssl.PKCS7_PADDING)
}

func DesECBEncryptPKCG5(src, key []byte) (string, error) {
	dst, err := openssl.DesECBEncrypt(src, key, openssl.PKCS5_PADDING)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(dst), nil
}

func DesECBDecryptPKCG5(src, key []byte) ([]byte, error) {
	decode, _ := base64.StdEncoding.DecodeString(string(src))
	return openssl.DesECBDecrypt(decode, key, openssl.PKCS5_PADDING)
}

func DesECBEncryptPKCG7(src, key []byte) (string, error) {
	dst, err := openssl.DesECBEncrypt(src, key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(dst), nil
}

func DesECBDecryptPKCG7(src, key []byte) ([]byte, error) {
	decode, _ := base64.StdEncoding.DecodeString(string(src))
	return openssl.DesECBDecrypt(decode, key, openssl.PKCS7_PADDING)
}

func DesECBEncrypt(src, key []byte) (string, error) {
	dst, err := openssl.DesECBEncrypt(src, key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(dst), nil
}

func DesECBDecrypt(src, key []byte) ([]byte, error) {
	decode, _ := base64.StdEncoding.DecodeString(string(src))
	return openssl.DesECBDecrypt(decode, key, openssl.PKCS7_PADDING)
}

func DesCBCEncrypt(src, key, iv []byte) (string, error) {
	dst, err := openssl.DesCBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(dst), nil
}

func DesCBCDecrypt(src, key, iv []byte) ([]byte, error) {
	decode, _ := base64.StdEncoding.DecodeString(string(src))
	return openssl.DesCBCDecrypt(decode, key, iv, openssl.PKCS7_PADDING)
}

func Des3ECBEncrypt(src, key []byte) (string, error) {
	dst, err := openssl.Des3ECBEncrypt(src, key, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(dst), nil
}

func Des3ECBDecrypt(src, key []byte) ([]byte, error) {
	decode, _ := base64.StdEncoding.DecodeString(string(src))
	return openssl.Des3ECBDecrypt(decode, key, openssl.PKCS7_PADDING)
}

func Des3CBCEncrypt(src, key, iv []byte) (string, error) {
	dst, err := openssl.Des3CBCEncrypt(src, key, iv, openssl.PKCS7_PADDING)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(dst), nil
}

func Des3CBCDecrypt(src, key, iv []byte) ([]byte, error) {
	decode, _ := base64.StdEncoding.DecodeString(string(src))
	return openssl.Des3CBCDecrypt(decode, key, iv, openssl.PKCS7_PADDING)
}

func RandStr(bit int) string {
	b := frand.Bytes(bit)
	rp := hex.EncodeToString(b)
	return rp
}

func AesEncrypt(plaintext []byte, key []byte, iv []byte) (string, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("err=", err)
		return "", err
	}
	blockSize := block.BlockSize()
	plaintext = PKCS5Padding(plaintext, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	blockMode.CryptBlocks(ciphertext, plaintext)
	encryptString := base64.RawURLEncoding.EncodeToString([]byte(ciphertext))
	return encryptString, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {

	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{
		byte(padding),
	}, padding)
	return append(ciphertext, padtext...)
}

func AesDecrypt(ciphertext string, key []byte, iv []byte) (string, error) {

	decode_data, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("err=", err)
		return "", err
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(decode_data))
	blockModel.CryptBlocks(plaintext, decode_data)
	plaintext = PKCS5UnPadding(plaintext)
	return string(plaintext), nil
}

func PKCS5UnPadding(ciphertext []byte) []byte {

	length := len(ciphertext)
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}

func AuthGetCode(fctx *fasthttp.RequestCtx, secret string) string {
	secret = strings.ToUpper(secret)

	timeSlice := fctx.Time().Unix() / 30

	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return ""
	}

	hash := hmac.New(sha1.New, key)
	err = binary.Write(hash, binary.BigEndian, timeSlice)
	if err != nil {
		return ""
	}
	h := hash.Sum(nil)

	offset := h[19] & 0x0f

	truncated := binary.BigEndian.Uint32(h[offset : offset+4])

	truncated &= 0x7fffffff
	code := truncated % 1000000

	//return int(code)
	return fmt.Sprintf("%06d", code)
}

// 兼容filbet之前密码加密方式
func BcryptEncrypt(p string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p), 10)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func BcryptVerify(s, d string) error {
	err := bcrypt.CompareHashAndPassword([]byte(s), []byte(d))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return fmt.Errorf("password mismatch")
		}
		return err
	}
	return nil
}
