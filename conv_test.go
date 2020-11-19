package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"corona/config"
	"corona/controller"
)

func TestConv(t *testing.T) {

	conf := config.Init()

	s := httptest.NewServer(&controller.ConvHandler{Conf: conf})
	defer s.Close()

	res, e := http.Get(s.URL + "/s")
	if e != nil {
		t.Errorf("Conv Get Test failed: %v", e)
		return
	}

	if !(res.StatusCode == http.StatusOK ||
		res.StatusCode == http.StatusNoContent) {
		t.Errorf("Conv Status Test failed: %s", res.Status)
		return
	}

}
