package controller

import (
	"net/http"

	"corona/config"
	"corona/model/click"
	"corona/model/conversion"
	"corona/util/tool"
)

type ConvHandler struct {
	Conf config.Obj
}

func (ch *ConvHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ck, rv, ev, tt := r.FormValue("c"), r.FormValue("r"), r.FormValue("e"), r.FormValue("t")

	d, c, e := tool.Split(ck, "_")
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	obj, ok := click.GetValid(ch.Conf, d, c)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if tt == "1" {
		w.Write([]byte("TEST OK"))
		return
	}

	if conversion.Exists(ch.Conf, c, ev) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	conv, ok := conversion.New(ch.Conf, obj, rv, ev)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	conv.Cap()

	conv.Save()

	if conv.Valid() {
		conv.Postback()
	}

	w.Write([]byte("OK"))
}
