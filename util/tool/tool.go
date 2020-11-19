package tool

import (
	"bytes"
	"encoding/base64"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func String(payload interface{}) string {
	var load string
	if pay, oh := payload.(string); oh {
		load = pay
	} else {
		load = ""
	}
	return load
}

func Float(payload interface{}) float64 {
	var load float64
	if pay, oh := payload.(float64); oh {
		load = pay
	} else {
		load = 0
	}
	return load
}

func Int(payload interface{}) int {
	var load int
	if pay, oh := payload.(int); oh {
		load = pay
	} else {
		load = 0
	}
	return load
}

type ipRange struct {
	start net.IP
	end   net.IP
}

// inRange - check to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
	// strcmp type byte comparison
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
		return true
	}
	return false
}

func isPrivateSubnet(ipAddress net.IP) bool {

	var privateRanges = []ipRange{
		{
			start: net.ParseIP("10.0.0.0"),
			end:   net.ParseIP("10.255.255.255"),
		},
		{
			start: net.ParseIP("100.64.0.0"),
			end:   net.ParseIP("100.127.255.255"),
		},
		{
			start: net.ParseIP("172.16.0.0"),
			end:   net.ParseIP("172.31.255.255"),
		},
		{
			start: net.ParseIP("192.0.0.0"),
			end:   net.ParseIP("192.0.0.255"),
		},
		{
			start: net.ParseIP("192.168.0.0"),
			end:   net.ParseIP("192.168.255.255"),
		},
		{
			start: net.ParseIP("198.18.0.0"),
			end:   net.ParseIP("198.19.255.255"),
		},
	}
	// my use case is only concerned with ipv4 atm
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		// iterate over all our ranges
		for _, r := range privateRanges {
			// check if this ip is in a private range
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}

func GetIPAddress(r *http.Request) (ip string) {
	ip = r.RemoteAddr
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		// march from right to left until we get a public address
		// that will be the address right before our proxy.
		for i := len(addresses) - 1; i >= 0; i-- {
			addr := strings.TrimSpace(addresses[i])
			// header can contain spaces too, strip those out.
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
				// bad address, go to next
				continue
			}
			ip = addr
			break
		}
	}
	ip = strings.Split(ip, ":")[0]
	return
}

func Get(this interface{}, key string) interface{} {
	return this.(map[interface{}]interface{})[key]
}

func InArray(val string, array []string) (exists bool) {
	exists = false

	for _, v := range array {
		if strings.EqualFold(v, val) {
			exists = true
			return
		}
	}
	return
}

func Day() int {
	day, _ := strconv.Atoi(time.Now().Format("20060102"))
	return day
}

func DayHour() int {
	day, _ := strconv.Atoi(time.Now().Format("2006010215"))
	return day
}

func Date() int {
	day, _ := strconv.Atoi(time.Now().Format("20060102150405"))
	return day
}

func Atob(a string) string {
	return base64.StdEncoding.EncodeToString([]byte(a))
}

func Btoa(b string) string {
	decoded, err := base64.StdEncoding.DecodeString(b)
	if err != nil {
		panic(err)
	}
	return string(decoded)
}

func Utob(a string) string {
	return base64.URLEncoding.EncodeToString([]byte(a))
}

func Btou(b string) string {
	decoded, err := base64.URLEncoding.DecodeString(b)
	if err != nil {
	}
	return string(decoded)
}

func Random(min, max int) int {
	return min + rand.Intn(max-min)
}

func Valid(ip, ua string) bool {

	if ip == "" || ua == "" {
		return false
	}

	ipv := net.ParseIP(ip)
	if ipv.To4() == nil && ipv.To16() == nil {
		return false
	}

	return true
}
func StrContains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}
	return false
}

func StrIntersects(str1 []string, str2 []string) bool {
	m := make(map[string]bool)
	for _, v := range str1 {
		m[v] = true
	}
	for _, v := range str2 {
		if m[v] {
			return true
		}
	}
	return false
}

func Split(s, d string) (string, string, error) {
	arr := strings.Split(s, d)
	if len(arr) < 2 {
		return s, "", errors.New("no delimiter")
	}

	return arr[0], arr[1], nil
}