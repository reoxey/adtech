package userdetect

import (
	"fmt"
	"strings"

	"github.com/ip2location/ip2location-go"
	FiftyOneDegreesTrieV3 "github.com/reoxey/Device-Detection-go/src/trie"
)

var provider FiftyOneDegreesTrieV3.Provider

type Device struct {
	Country string
	City    string
	Region  string
	OS      string
	Version string
	DT      string
	Browser string
	IsBot   bool
	IP      string
	UA      string
	ISP     string
	Carrier string
	Brand   string
}

func Init(d51, ip2l string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from 51D panic")
		}
	}()
	provider = FiftyOneDegreesTrieV3.NewProvider(d51)

	ip2location.Open(ip2l)
}

func Match(ip, ua, city, geo, os, ver string) (Device, bool) {
	//get := detect(ip, ua)
	get := detect("106.210.143.160",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 6_1_4 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Mobile/10B350")

	fmt.Println(get, os, geo)

	ok := true
	if city != "" && city != get.City {
		ok = false
	} else if os != get.OS && geo != get.Country {
		ok = false
	}
	return get, ok
}

func detect(ip, ua string) (all Device) {
	all = device(ua)
	geo := ip2location.Get_all(ip)

	all.Country = geo.Country_short
	all.Region = geo.Region
	all.City = geo.City
	all.ISP = geo.Isp
	all.Carrier = geo.Mobilebrand

	all.IP = ip
	all.UA = ua

	return
}

func device(ua string) Device {

	if provider == nil {
		return Device{}
	}

	match := provider.GetMatch(ua)

	device := Device{
		IsBot:   match.GetValue("IsCrawler") == "true",
		OS:      match.GetValue("PlatformName"),
		Version: match.GetValue("PlatformVersion"),
		DT:      match.GetValue("DeviceType"),
		Brand:   match.GetValue("HardwareVendor"),
		Browser: match.GetValue("BrowserName"),
	}
	FiftyOneDegreesTrieV3.DeleteMatch(match)

	if device.OS == "Unknown" {
		if strings.Index(ua, "like Mac OS X") > -1 {
			device.OS = "iOS"
		} else if strings.Index(ua, "Android") > -1 {
			device.OS = "Android"
		}
	}

	return device
}
