package main

import (
	"fmt"
	aufs "github.com/aulaga/aufs/src"
	"github.com/aulaga/aufs/src/webdav"
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

	r.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = aufs.Context(ctx, &SampleFs{})
			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
		})
	})

	r.Mount("/dav", webdav.Handler())

	err := http.ListenAndServe("0.0.0.0:8080", r)
	if err != nil {
		panic(err.Error())
	}
}
