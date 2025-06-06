package helper

import (
	"encoding/binary"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
)

// Should use canonical format of the header key s
// https://golang.org/pkg/net/http/#CanonicalHeaderKey

// Header may return multiple IP addresses in the format: "client IP, proxy 1 IP, proxy 2 IP", so we take the the first one.
var xForwardedForHeader = http.CanonicalHeaderKey("X-Original-Forwarded-For")
var xForwardedHeader = http.CanonicalHeaderKey("X-Forwarded-For")
var forwardedForHeader = http.CanonicalHeaderKey("Forwarded-For")
var forwardedHeader = http.CanonicalHeaderKey("Forwarded")

// Standard headers used by Amazon EC2, Heroku, and others
var xClientIPHeader = http.CanonicalHeaderKey("X-Forwarded-For")

// Nginx proxy/FastCGI
var xRealIPHeader = http.CanonicalHeaderKey("X-Real-IP")

/*
// Cloudflare.
// @see https://support.cloudflare.com/hc/en-us/articles/200170986-How-does-Cloudflare-handle-HTTP-Request-headers-
// CF-Connecting-IP - applied to every request to the origin.
var cfConnectingIPHeader = http.CanonicalHeaderKey("X-Original-Forwarded-For")

// Fastly CDN and Firebase hosting header when forwared to a cloud function
var fastlyClientIPHeader = http.CanonicalHeaderKey("Fastly-Client-Ip")

// Akamai and Cloudflare
var trueClientIPHeader = http.CanonicalHeaderKey("True-Client-Ip")
*/
var cidrs []*net.IPNet

func Ip2long(ipAddr string) (uint32, error) {

	if !IsIPv4(ipAddr) {
		return 0, nil
	}

	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return 0, errors.New("wrong ipAddr format")
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip), nil
}

// 判断是否是合法的ipv4
func IsIPv4(address string) bool {
	return net.ParseIP(address).To4() != nil
}

// 判断是否是合法的ipv6
func IsIPv6(address string) bool {
	return net.ParseIP(address).To16() != nil
}

// 判断是否是合法的ip
func IsIP(address string) bool {
	return IsIPv4(address) || IsIPv6(address)
}

func init() {
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}

	cidrs = make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		cidrs[i] = cidr
	}
}

// isLocalAddress works by checking if the address is under private CIDR blocks.
// List of private CIDR blocks can be seen on :
//
// https://en.wikipedia.org/wiki/Private_network
//
// https://en.wikipedia.org/wiki/Link-local_address
func isPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, errors.New("address is not valid")
	}

	for i := range cidrs {
		if cidrs[i].Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}

// FromRequest returns client's real public IP address from http request headers.
func FromRequest(ctx *fasthttp.RequestCtx) string {
	/*
			  	xClientIP := ctx.Request.Header.Peek(xClientIPHeader)
				if xClientIP != nil {
					return string(xClientIP)
				}

				xForwardedFor := ctx.Request.Header.Peek(xForwardedForHeader)
				if xForwardedFor != nil {
					requestIP, err := retrieveForwardedIP(string(xForwardedFor))
					if err == nil {
						return requestIP
					}
				}


		if ip, err := fromForwardedHeaders(ctx); err == nil {

			requestIP, err := retrieveForwardedIP(ip)
			if err == nil {
				return requestIP
			}
			//return ip
		}
	*/
	if ip, err := fromSpecialHeaders(ctx); err == nil {

		requestIP, err := retrieveForwardedIP(ip)
		if err == nil {
			return requestIP
		}
		//return ip
	}

	var remoteIP string
	remoteAddr := ctx.RemoteAddr().String()

	if strings.ContainsRune(remoteAddr, ':') {
		remoteIP, _, _ = net.SplitHostPort(remoteAddr)
	} else {
		remoteIP = remoteAddr
	}
	return remoteIP
}

func fromSpecialHeaders(ctx *fasthttp.RequestCtx) (string, error) {

	if clientIP := ctx.Request.Header.Peek(xRealIPHeader); clientIP != nil {
		return string(clientIP), nil
	}
	return "", errors.New("can't get ip from special headers")
}

func fromForwardedHeaders(ctx *fasthttp.RequestCtx) (string, error) {
	forwardedHeaders := [...]string{xForwardedHeader, forwardedForHeader, forwardedHeader}
	for _, forwardedHeader := range forwardedHeaders {
		if forwarded := ctx.Request.Header.Peek(forwardedHeader); forwarded != nil {
			if clientIP, err := retrieveForwardedIP(string(forwarded)); err == nil {
				return clientIP, nil
			}
		}
	}
	return "", errors.New("can't get ip from forwarded headers")
}

func retrieveForwardedIP(forwardedHeader string) (string, error) {
	for _, address := range strings.Split(forwardedHeader, ",") {
		if len(address) > 0 {
			address = strings.TrimSpace(address)
			isPrivate, err := isPrivateAddress(address)
			switch {
			case !isPrivate && err == nil:
				return address, nil
			case isPrivate && err == nil:
				return "", errors.New("forwarded ip is private")
			default:
				return "", err
			}
		}
	}
	return "", errors.New("empty or invalid forwarded header")
}
