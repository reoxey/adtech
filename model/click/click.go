package click

import (
	"strconv"
	"strings"
	"time"

	"corona/config"
	"corona/model/offer"
	"corona/util/tool"
	"corona/util/userdetect"
)

// Tables names
const CLICK = "click_"
const clickFILTER = "click_filter"

// clickDROP to flag click if invalid
type clickDROP int
const (
	NonTARGET clickDROP = 11
	IsFILTERED clickDROP = 22
	IsCAPPED clickDROP = 33
	IsDUPLICATE clickDROP = 44
	MissingREQUIRED clickDROP = 55
)

// Click details to be saved in Cassandra for stats
type Click struct {
	Id        string
	Oid       int
	Pid       int
	Nid       int
	Offer     string
	Publisher string
	Bundle	  string
	Category  string
	Geo       string
	Region    string
	City      string
	Os        string
	Version   string
	Dt        string
	Browser   string
	Ip        string
	Ua        string
	Isp       string
	Carrier   string
	Brand     string
	Ifa       string
	App       string
	Trkid     string
	Subid     string
	Day       int
	Hour      int
	Datetime  int
	Hit		  int
	Miss      int
	Misscode  clickDROP
}

// Env model specific data
type Env struct {
	pub, off, adv   string
	conf config.Obj
	*Click
}

// Clicker implements Env
type Clicker interface {
	Filter(string) bool
	SaveValid(...string)
	SaveInvalid(clickDROP, ...string)
	Target(string, string, string) bool
}

var _ Clicker = (*Env) (nil)

// New initialise Click with details
func New(o *offer.Env, c, ifa, app string) Clicker {
	oid, _ := strconv.Atoi(o.Oid)
	pid, _ := strconv.Atoi(o.Pid)
	nid, _ := strconv.Atoi(o.Nid)

	clk := &Click{
		Oid: oid,
		Pid: pid,
		Nid: nid,
		Offer: o.Off,
		Publisher: o.Pub,
		Bundle: o.Pkg,
		Os: o.OS,
		Category: o.Cat,
		Geo: o.Geo,
		City: o.City,
		Trkid: c,
		Ifa: ifa,
		App: app,
	}

	return Env{
		pub:   o.Pid,
		off:   o.Oid,
		adv:   o.Nid,
		conf:  o.Conf,
		Click: clk,
	}
}

// Filter subid from publisher based on database entries
func (en Env) Filter(s string) (ok bool) {
	//TODO filter
	en.Subid = s

	if s == "" {
		return
	}

	k := "FIL::"+s+":"+en.pub+":"+en.adv
	offers, e := en.conf.RD.Get(k)

	if e != nil && len(offers) == 0 {

		conn, e := en.conf.DB.Open()
		if e != nil {
			return false
		}
		defer en.conf.DB.Close()

		o,e := conn.Select("GROUP_CONCAT(offer_id)").Table(clickFILTER).
			Run("ftr_sub_id = '"+s+"'"+" AND publisher_id = "+en.pub +" AND advertiser_id = "+en.adv)
		for o.Next() {
			offers = o.String()
			en.conf.RD.SetEx(k, offers, 30*time.Minute)
		}
	}

	switch {
	case offers == "X":
	case offers == "":
		en.conf.RD.SetEx(k, "X", 30*time.Minute)
	case strings.Contains(offers, en.off) && offers == "O":
		ok = true
	}

	return
}

// SaveValid if Click with clickDROP is not flagged
func (en Env) SaveValid(ex ...string) {

	en.Hit = 1

	go insert(en.conf, en.Click)
}

// SaveInvalid if Click with clickDROP is flagged
func (en Env) SaveInvalid(code clickDROP, ex ...string) {

	en.Miss = 1
	en.Misscode = code

	go insert(en.conf, en.Click)
}

// insert into Cassandra
func insert(conf config.Obj, row *Click) {
	if conf.QL.Session != nil {

		row.Id = conf.QL.UUID()
		row.Day = tool.Day()
		row.Hour = time.Now().Hour()
		row.Datetime = int(time.Now().Unix())
		// conf.QL.Put(CLICK+strconv.Itoa(row.Day)+strconv.Itoa(row.Hour), row)
		conf.QL.Put(CLICK, row)
	}
}

// GetValid pulls valid Click at the time sale
func GetValid(conf config.Obj, d, c string) (clk Click, ok bool) {

	e := conf.QL.PullOne(CLICK, &clk, map[string]interface{}{
		"id": c,
	},
		conf.QL.W("eq", "id"),
	)

	if e != nil {
		return
	}

	return clk, true
}

// Target Click to the specified targeting details of the offer.Campaign.
func (en Env) Target(ip, ua, v string) (ok bool) {

	dev, ok := userdetect.Match(ip, ua, en.City, en.Geo, en.Os, v)

	en.Os = dev.OS
	en.City = dev.City
	en.Geo = dev.Country
	en.Isp = dev.ISP
	en.Region = dev.Region
	en.Brand = dev.Brand
	en.Ip = dev.IP
	en.Ua = dev.UA
	en.Browser = dev.Browser
	en.Carrier = dev.Carrier
	en.Dt = dev.DT
	en.Version = dev.Version

	return
}