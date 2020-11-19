package conversion

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"corona/config"
	"corona/model/click"
)

const (
	defaultEVENT = "install"
	REVSHARE = "REVSHARE"
	FIXED = "FIXED"
)

// Tables names
const (
	ASSIGNED = "assigned"
	OFFER = "offer"
 	PUBLISHER = "publisher"
 	CONVERSION = "conversion"
)

// dropCODE to flag conversion if invalid
type dropCODE string
const (
	mismatchGEO dropCODE = "111"
	mismatchOS dropCODE = "222"
	noEVENT dropCODE = "333"
	capPubREACHED dropCODE = "444"
	capOffREACHED dropCODE = "555"
)

// Conv details to be save in mysql for each conversion
type Conv struct {
	Event string
	click.Click
	Revenue, Payout float32
	Amount string
	Drop, Reason string
	DropCode dropCODE
	Paid, Pay string
}

// Env model specific data
type Env struct {
	pub, off   string
	conf config.Obj
	*Conv
	access int
}

// Convert implements Env
type Convert interface {
	Cap()
	Save()
	Valid() bool
	Postback()
}

var _ Convert = (*Env) (nil)

// New initialise Conv with conversion details
func New(conf config.Obj, c click.Click, rv, ev string) (Convert, bool) {

	if ev == "" {
		ev = defaultEVENT
	}

	ok := false

	d := &Conv{Amount: rv, Click: c, Event: ev, Drop: "0",
		Pay: "0", Paid: "0" }

	en := Env{
		conf: conf,
		pub:    strconv.Itoa(c.Pid),
		off:    strconv.Itoa(c.Oid),
		Conv:   d,
		access: 0,
	}

	conn, e := conf.DB.Open()
	if e != nil{
		return nil, false
	}
	defer conf.DB.Close()

	o, e := conn.Select("off_access", "off_revenue", "off_payout", "events", "off_default_event").
		Table(OFFER).Run("off_id = "+en.off)
	if e != nil{
		return nil, false
	}

	for o.Next() {

		ok = true

		en.access = o.Int()
		d.Revenue = o.Float32()
		d.Payout = o.Float32()

		evs := o.String()
		de := o.String()

		switch {
		case de == ev: d.Paid = "1"
		case evs != "[]" && d.events(evs): d.Paid = "1"
		default:
			d.dropped(noEVENT, ev + " is not specified. Default Event "+de)
		}
	}

	return en, ok
}

// Exists check if conversion is already been recorded in DB
func Exists(conf config.Obj, c, ev string) (ok bool) {

	conn, e := conf.DB.Open()
	if e != nil{
		return
	}
	defer conf.DB.Close()

	o, e := conn.Select("con_click").Table(CONVERSION).
		Run("con_click = "+c +" AND con_event = "+ev)

	if e != nil{
		return
	}

	for o.Next() {
		ok = true
	}

	return
}

// events parse JSON saved in the DB to get details about event details
func (c *Conv) events(s string) (ok bool) {

	type eve struct {
		E string `json:"e"`
		R float32 `json:"r"`
		P float32 `json:"p"`
	}
	var ev []eve
	if json.Unmarshal([]byte(s), &ev) != nil {
		return
	}

	for _, v := range ev {
		if v.E == c.Event {
			c.Revenue = v.R
			c.Payout = v.P
			return true
		}
	}

	return
}

// Cap check and update Cap of the offer and publisher if conversion is valid.
func (en Env) Cap() {

	if en.Drop == "1" {
		return //TODO to be clarify. cap on non event.
	}

	if count, ok := isPubCapped(en.conf, en.off, en.pub); ok {
		en.dropped(capPubREACHED, "Cap Pub reached: "+count)
	}

	if count, ok := isOffCapped(en.conf, en.off); ok {
		en.dropped(capOffREACHED, "Cap Off reached: "+count)
	}

}

// Save conversion is stored in DB
func (en Env) Save()  {
	conn, e := en.conf.DB.Open()
	if e != nil{
		return
	}
	defer en.conf.DB.Close()

	now := int(time.Now().Unix())
	ctit := now - en.Datetime

	m := map[string]string{
		"offer_name":     en.Offer,
		"publisher_name": en.Publisher,
		"con_category":   en.Category,
		"con_bundle":     en.Bundle,
		"con_geo":        en.Geo,
		"con_region":     en.Region,
		"con_city":       en.City,
		"con_isp":        en.Isp,
		"con_carrier":    en.Carrier,
		"con_os":         en.Os,
		"con_ip":         en.Ip,
		"con_ua":         en.Ua,
		"con_osv":        en.Version,
		"con_dt":         en.Dt,
		"con_brand":      en.Brand,
		"con_browser":    en.Browser,
		"con_click_time": strconv.Itoa(en.Datetime),
		"con_click_day":  strconv.Itoa(en.Day),
		"con_click_hour": strconv.Itoa(en.Hour),
		"con_time":       strconv.Itoa(now),
		"con_day":        time.Now().Format("20060102"),
		"con_hour":       strconv.Itoa(time.Now().Hour()),
		"con_ctit":       strconv.Itoa(ctit),
		"con_revenue":    strconv.FormatFloat(float64(en.Revenue), 'f', -1, 32),
		"con_payout":     strconv.FormatFloat(float64(en.Payout), 'f', -1, 32),
		"con_ifa":        en.Ifa,
		"con_app":        en.App,
		"con_click":      en.Trkid,
		"con_source":     en.Subid,
		"con_event":      en.Event,
		"con_amount":     en.Amount,
		"con_paid":       en.Paid,
		"con_pay":        en.Pay,
		"offer_id":       en.off,
		"publihser_id":   en.pub,
		"advertiser_id":  strconv.Itoa(en.Nid),
	}

	if e := conn.Set(m).Table(CONVERSION).Put(); e != nil {
		//TODO log
	}
}

// dropped if invalid
func (c Conv) dropped(code dropCODE, reason string)  {
	c.Drop = "1"
	c.Paid = "0"
	c.Pay = "0"
	c.DropCode = code
	c.Reason = reason
}

// Valid if Drop is zero and Pay is one
func (en Env) Valid() bool {

	if en.Drop == "0" && en.Pay == "1" {
		return true
	}

	return false
}

// Postback is fried to publishers if Pay and event matched per specified.
func (en Env) Postback() {

	conn, e := en.conf.DB.Open()
	if e != nil{
		return
	}
	defer en.conf.DB.Close()

	var pCallback string

	if en.access == 0 {
		o, e := conn.Select( "pub_type", "pub_share", "pub_callback").
			Table(PUBLISHER).Run("pub_id = " + en.pub + " AND pub_status = 1")
		if e != nil{
			return
		}

		for o.Next() {
			pType := o.String()
			pShare := o.Float32()
			pCallback = o.String()

			if pType == REVSHARE {
				en.Payout = en.Revenue * pShare
			}
		}
	} else {
		o, e := conn.Select( "ass_type", "ass_share", "pub_callback", "events").
			Table(ASSIGNED).Join(PUBLISHER).
			Run("assigned.off_id = " + en.off + " AND assigned.pub_id = " + en.pub + " AND publisher.pub_id = assigned.pub_id AND pub_status = 1 AND ass_status = 1")
		if e != nil{
			return
		}

		for o.Next() {
			aType := o.String()
			aShare := o.Float32()
			pCallback = o.String()
			events := o.String()

			if en.events(events) {
				if aType == REVSHARE {
					en.Payout = en.Revenue * aShare
				}
				en.Pay = "1"
			}
		}
	}

	if pCallback == "" || en.Pay == "0" {
		//TODO log
		return
	}

	go tryFire(pCallback, 1)
}

// tryFire retries 3 times for non 2xx responses
func tryFire(url string, x int)  {
	if x < 4 {
		x++
		r, e := http.Get(url)
		if e != nil {
			tryFire(url, x)
			//TODO log

		} else if !(r.StatusCode == 200 || r.StatusCode == 204) {
			tryFire(url, x)
		}
	}
}

func isPubCapped(conf config.Obj, oid, pid string) (x string, ok bool) {
	conn, e := conf.DB.Open()
	if e != nil{
		return
	}
	defer conf.DB.Close()

	count := 0
	date := time.Now().Format("20060102")

	obj, e := conn.Select("ass_cap_limit", "ass_cap_count").Table(ASSIGNED).
		Run("pub_id = "+pid +" AND off_id = "+oid + " AND  AND ass_cap_date = " + date)
	if e != nil{
		return
	}

	for obj.Next() {
		cp := obj.Int()
		count = obj.Int()

		if cp > 0 && cp <= count {
			ok = true
		}
	}

	count++
	x = strconv.Itoa(count)
	if conn.Set(map[string]string{
		"ass_cap_count": x,
		"ass_cap_date" : date,
	}).
		Table(ASSIGNED).
		Update("off_id = "+oid +" AND pub_id = "+pid) != nil {
			//TODO log the error
	}

	return
}

func isOffCapped(conf config.Obj, oid string) (x string, ok bool) {
	conn, e := conf.DB.Open()
	if e != nil{
		return
	}
	defer conf.DB.Close()

	count := 0
	date := time.Now().Format("20060102")

	obj, e := conn.Select("off_conv_cap", "off_conv_count").Table(OFFER).
		Run("off_id = "+oid + " AND off_cap_date = " + date)
	if e != nil{
		return
	}

	for obj.Next() {
		cp := obj.Int()
		count = obj.Int()

		if cp > 0 && cp <= count {
			ok = true
		}
	}

	count++
	x = strconv.Itoa(count)
	if conn.Set(map[string]string{
		"ass_cap_count": x,
		"ass_cap_date" : date,
	}).
		Table(OFFER).
		Update("off_id = "+oid) != nil {
		//TODO log the error
	}

	return
}