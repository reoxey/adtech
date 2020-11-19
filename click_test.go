package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"corona/config"
	"corona/controller"
)

func TestClick(t *testing.T) {

	conf := config.Init()

	s := httptest.NewServer(&controller.ClickHandler{Conf: conf})
	defer s.Close()

	res, e := http.Get(s.URL + "/c")
	if e != nil {
		t.Errorf("Click Get Test failed: %v", e)
		return
	}

	if !(res.StatusCode == http.StatusOK ||
		res.StatusCode == http.StatusTemporaryRedirect ||
		res.StatusCode == http.StatusFound) {
		t.Errorf("Click Status Test failed: %s", res.Status)
		return
	}

}
