package main

import (
	"fmt"
	"github.com/aulaga/cloud/src/webdav"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	fmt.Println("Starting...")

	chi.RegisterMethod("PROPFIND")
	chi.RegisterMethod("PROPPATCH")
	chi.RegisterMethod("MKCOL")
	chi.RegisterMethod("COPY")
	chi.RegisterMethod("MOVE")
	chi.RegisterMethod("LOCK")
	chi.RegisterMethod("UNLOCK")
	r := chi.NewRouter()

	r.Mount("/dav", webdav.Handler())

	err := http.ListenAndServe("0.0.0.0:8080", r)
	if err != nil {
		panic(err.Error())
	}
}
