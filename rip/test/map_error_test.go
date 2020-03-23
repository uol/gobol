package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/uol/gobol/rip"
	"github.com/uol/gobol/snitch"
)

func TestMapError(t *testing.T) {

	createServer()

	// testMapErrorCodeless()

}

// func TestMapErrorCodeless(t *testing.T) {
// 	fmt.Println("starting testMapErrorCodeless")

// 	resp, err := http.Get("http://0.0.0.0:8080/probe")
// 	if err != nil {
// 		fmt.Println("HTTP request failed with error", err)
// 	} else {
// 		data, _ := ioutil.ReadAll(resp.Body)
// 		fmt.Println("data", string(data))
// 	}
// }

func createServer() {
	// create stats
	statsSettings := snitch.Settings{
		Address:          "a1-labirinto4.host.intranet",
		Port:             80,
		Protocol:         "http",
		HTTPTimeout:      "10s",
		HTTPPostInterval: "10s",
		Tags: map[string]string{
			"ttl":     "1",
			"ksid":    "pdeng_collectd",
			"service": "riptest",
		},
		Interval: "@every 1m",
		Runtime:  true,
	}

	stats, err := snitch.New(statsSettings)
	if err != nil {
		fmt.Println("error creating stats", err)
	}

	// create server with probeTest endpoint (http://0.0.0.0:8080/probe) that Fail all request
	router := rip.NewCustomRouterMapError("errors.json")
	router.GET("/probe", probeTest)

	server := &http.Server{
		Addr: "0.0.0.0" + ":" + "8088",
		Handler: rip.NewLogMiddleware(
			"chimera",
			"macs",
			stats,
			rip.NewGzipMiddleware(rip.BestSpeed, router),
			false,
		),
	}

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Println("error starting server", err)
		return
	}
}

// endpoint that fails all request
func probeTest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	msg := "teste de mensagem de gerr"
	gerr := mError{
		fmt.Errorf("erro: %s", msg),
		http.StatusBadRequest,
		"mensagem: " + msg,
		"test",
		"probeTest",
		"cod1",
	}
	rip.Fail(w, gerr)
	return
}

type mError struct {
	error
	httpCode int
	msg      string
	pkg      string
	function string
	code     string
}

func (e mError) Message() string {
	return e.msg
}

func (e mError) Package() string {
	return e.pkg
}

func (e mError) Function() string {
	return e.function
}

func (e mError) StatusCode() int {
	return e.httpCode
}

func (e mError) Error() string {
	return e.error.Error()
}

func (e mError) ErrorCode() string {
	return e.code
}
