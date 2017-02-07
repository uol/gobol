package rip

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func NewCustomRouter() *httprouter.Router {

	router := httprouter.New()
	router.MethodNotAllowed = customNotAllowed{}
	router.NotFound = customNotFound{}
	return router
}

type customNotFound struct{}

func (cnf customNotFound) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

type customNotAllowed struct{}

func (cna customNotAllowed) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
