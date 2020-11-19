package controller

import (
	"net/http"

	"corona/config"
)

// Attach add routes controller
func Attach(conf config.Obj) *http.ServeMux {
	mux := http.NewServeMux()

	clickHandler := &ClickHandler{conf}
	mux.Handle("/c", clickHandler)

	convHandler := &ConvHandler{conf}
	mux.Handle("/s", convHandler)

	return mux
}
