package offer

import (
	"encoding/json"
	"time"

	"corona/config"
)

// Tables names
const ASSIGNED = "assigned"
const OFFER = "offer"
const PUBLISHER = "publisher"

// Env model specific data
type Env struct {
	Pid, Oid, Nid   string
	Conf config.Obj
	*Campaign
}

// RevType Revenue Type (FIXED, REVSHARE)
type RevType string

// Campaign data cached in the redis
type Campaign struct {
	Off      string
	Pub		 string
	URL      string
	Pkg      string
	OS       string
	Geo      string
	City     string
	Access   int
	Cat		 string
	Type     RevType
}

// Op implements Env
type Op interface {
	Assigned() bool
}

var _ Op = (*Env) (nil)

// Get initialised object of *Env
func Get(conf config.Obj, p, o, n string) *Env {
	return &Env{
		Conf:       conf,
		Oid:        o,
		Pid:        p,
		Nid:        n,
	}
}

// Assigned validates the offer details and if the offer is assigned
// to publisher in case of private
func (en *Env) Assigned() (ok bool) {
	k := "PO::" + en.Pid + "::" + en.Oid

	var camp Campaign

	if cached, e := en.Conf.RD.Get(k); e != nil && len(cached) == 0 {

		if !en.setCamp() {
			println("1")
			en.Conf.RD.SetEx(k, "-", 15*time.Minute) //TODO time
			return
		}

		if cac, e := json.Marshal(camp); e != nil {
			println("3")
			return
		} else {
			en.Conf.RD.SetEx(k, string(cac), 5*time.Minute)
		}
		

	} else {

		if cached == "-" {
			return
		}

		if e := json.Unmarshal([]byte(cached), &camp); e != nil {
			println("4")
			return
		}
	}

	en.Campaign = &camp

	return true
}

// setCamp updates the redis cache with Campaign details
func (en Env) setCamp() bool {

	conn, e := en.Conf.DB.Open()
	if e != nil {
		return false
	}
	defer en.Conf.DB.Close()

	of, e := conn.Select( "off_name", "off_bundle", "off_url", "off_os", "off_geo", "off_city", "off_access", "category").
		Table(OFFER).Run("off_id = " + en.Oid + " AND off_status = 1")
			//"((off_conv_count = 0 AND off_cap_date != '" + time.Now().Format("2006-01-02") + "') OR off_conv_count > 0)")
	if e != nil {
		return false
	}

	var camp Campaign

	for of.Next() {
		camp = Campaign{
			Off:	  of.String(),
			Pkg:	  of.String(),
			URL:      of.String(),
			OS:       of.String(),
			Geo:      of.String(),
			City:     of.String(),
			Access:   of.Int(),
			Cat: 	  of.String(),
		}
	}

	if camp.URL == "" {
		return false
	}

	if camp.Access == 0 {
		of, e = conn.Select( "pub_name", "pub_type").
			Table(PUBLISHER).Run("pub_id = " + en.Pid + " AND pub_status = 1")
	} else {
		of, e = conn.Select("pub_name", "ass_type").
			Table(ASSIGNED).Join(PUBLISHER).
			Run("assigned.off_id = " + en.Oid + " AND assigned.pub_id = " + en.Pid + " AND publisher.pub_id = assigned.pub_id AND pub_status = 1 AND ass_status = 1")
		//" AND ((ass_cap_count = 0 AND ass_cap_date != '" + time.Now().Format("2006-01-02") + "') OR ass_cap_count > 0)")
	}
	if e != nil {
		return false
	}

	for of.Next() {
		en.Pub = of.String()
		en.Type = RevType(of.String())
	}

	en.Campaign = &camp

	return true
}

// Capped weather publisher or offer is capped
func (en *Env) Capped() bool {

	return isOffCapped(en.Conf, en.Oid) && isPubCapped(en.Conf, en.Pid, en.Oid)
}

func isOffCapped(conf config.Obj, oid string) (ok bool) {
	k := "CAP::OFF::" + oid

	if campaign, e := conf.RD.Get(k); e != nil && len(campaign) == 0 {
		conn, e := conf.DB.Open()
		if e != nil {
			return false
		}
		defer conf.DB.Close()

		obj, e := conn.Select("off_conv_cap", "off_conv_count").Table(OFFER).
			Run("off_id = "+oid + " AND off_cap_date = " + time.Now().Format("20060102"))
		if e != nil {
			return false
		}

		for obj.Next() {
			cp := obj.Int()
			count := obj.Int()

			if cp > 0 && cp <= count {
				ok = true
				conf.RD.SetEx(k, "T", 5 * time.Minute)
			}
		}

		if !ok {
			conf.RD.SetEx(k, "F", 5 * time.Minute)
			return
		}
	} else {
		return campaign == "T"
	}

	return
}

func isPubCapped(conf config.Obj, pid, oid string) (ok bool) {
	k := "CAP::PUB::" + pid

	if campaign, e := conf.RD.Get(k); e != nil && len(campaign) == 0 {
		conn, e := conf.DB.Open()
		if e != nil {
			return false
		}
		defer conf.DB.Close()

		obj, e := conn.Select("ass_cap_limit", "ass_cap_count").Table(ASSIGNED).
			Run("pub_id = "+pid +" AND off_id = "+oid + " AND  AND ass_cap_date = " + time.Now().Format("20060102"))
		if e != nil {
			return false
		}

		for obj.Next() {
			cp := obj.Int()
			count := obj.Int()

			if cp > 0 && cp <= count {
				ok = true
				conf.RD.SetEx(k, "T", 5 * time.Minute)
			}
		}

		if !ok {
			conf.RD.SetEx(k, "F", 5 * time.Minute)
			return
		}
	} else {
		return campaign == "T"
	}

	return
}