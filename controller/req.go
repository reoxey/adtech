package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"corona/config"
	"corona/model/click"
	"corona/model/fallback"
	"corona/model/offer"
	"corona/util/tool"
)

const cooKIE = "xxx"
const clickID = "{clk}"
const sourceID = "{aid}"
const ifA = "{ifa}"
const appNAME = "{app}"

type ClickHandler struct {
	Conf config.Obj
}

func (ch *ClickHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	p := r.FormValue("p")     // publisher
	o := r.FormValue("o")     // offer
	n := r.FormValue("n")     // network
	s := r.FormValue("s")     // source id
	c := r.FormValue("c")     // click
	a := r.FormValue("a")     // app required flag 0-1
	app := r.FormValue("app") // app value
	f := r.FormValue("f")     // ifa required flag 0-1
	ifa := r.FormValue("ifa") // ifa value
	v := r.FormValue("v")     // os version
	ex1 := r.FormValue("ex1") // extra
	ex2 := r.FormValue("ex2")
	ex3 := r.FormValue("ex3")

	ip := tool.GetIPAddress(r)
	ua := r.UserAgent()

	var logs strings.Builder
	logs.WriteString(p + " | " + o + " | " + n + " | ") //TODO remove

	fb := fallback.Init(w, r.FormValue("fa"), r.FormValue("fx"), logs)
	// ?p=1000&o=1000000&n=2&s=123123&c=qwerty&a=0&app=&f=0&ifa=&v=&fa=0&fx=

	if p == "" && o == "" {
		fb.Redirect("F1 | ")
		return
	}

	off := offer.Get(ch.Conf, p, o, n)

	if !off.Assigned() {
		fb.Redirect("F2 | ")
		return
	}

	clk := click.New(off, c, ifa, app)

	// TODO use context with goroutines
	if off.Capped() {
		clk.SaveInvalid(click.IsCAPPED, ex1, ex2, ex3)
		fb.Redirect("F4 | ")
		return
	}

	if !clk.Target(ip, ua, v) {
		clk.SaveInvalid(click.NonTARGET, ex1, ex2, ex3)
		fb.Redirect("F3 | ")
		return
	}

	if clk.Filter(s) {
		clk.SaveInvalid(click.IsFILTERED, ex1, ex2, ex3)
		fb.Redirect("F5 | ")
		return
	}

	if isDuplicate(w, r) {
		clk.SaveInvalid(click.IsDUPLICATE, ex1, ex2, ex3)
		fb.Redirect("F6 | ")
		return
	}

	if f == "1" && ifa == "" || a == "1" && app == "" {
		clk.SaveInvalid(click.MissingREQUIRED, ex1, ex2, ex3)
		fb.Redirect("F7 | ")
		return
	}

	clk.SaveValid(ex1, ex2, ex3)

	url := replace(off.URL, c, p+"_"+s, ifa, app) //TODO source id
	fmt.Println(url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	fmt.Println(logs.String(), time.Since(start).String())
}

// isDuplicate checks the cookie validity for 60s
func isDuplicate(w http.ResponseWriter, r *http.Request) bool {
	_, e := r.Cookie(cooKIE)
	if e != nil {
		http.SetCookie(w, &http.Cookie{Name: cooKIE, Value: "__", Expires: time.Now().Add(time.Minute), HttpOnly: true, Path: "/"})
		return false
	}
	return true
}

func replace(url, c, s, ifa, app string) string {
	url = strings.Replace(url, clickID, c, 1)
	url = strings.Replace(url, sourceID, s, 1)
	url = strings.Replace(url, ifA, ifa, 1)
	return strings.Replace(url, appNAME, app, 1)
}
