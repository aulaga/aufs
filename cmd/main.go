package main

import (
	"encoding/json"
	"fmt"
	"github.com/aulaga/cloud/src/domain/storage"
	"github.com/aulaga/cloud/src/webdav"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
)

type BaseHandler struct {
	webdavHandler http.Handler
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	fmt.Println("respondWithJson", payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

type NextcloudStatus struct {
	Installed       bool   `json:"installed"`
	Maintenance     bool   `json:"maintenance"`
	NeedsDbUpgrade  bool   `json:"needsDbUpgrade"`
	Version         string `json:"version"`
	VersionString   string `json:"versionstring"`
	Edition         string `json:"edition"`
	ProductName     string `json:"productname"`
	ExtendedSupport bool   `json:"extendedSupport"`
}

type NextcloudPollEndpoint struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

type NextcloudLoginV2 struct {
	Poll  NextcloudPollEndpoint `json:"poll"`
	Login string                `json:"login"`
}

type NextcloudPollAuth struct {
	Server      string `json:"server"`
	LoginName   string `json:"loginName"`
	AppPassword string `json:"appPassword"`
}

func (h BaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	webdav.DebugRequest(w, r)

	// TODO handle basic auth
	for k, v := range r.Header {
		fmt.Println(k, v)
	}

	type empty struct{}

	username, password, hasBasicAuth := r.BasicAuth()
	correctLogin := username == "admin" && password == "admin"
	if hasBasicAuth && !correctLogin {
		respondWithJSON(w, http.StatusUnauthorized, empty{})
	}

	if r.URL.Path == "/status.php" {
		response := NextcloudStatus{
			Installed:       true,
			Maintenance:     false,
			NeedsDbUpgrade:  false,
			Version:         "20.0.0",
			VersionString:   "20.0.0",
			Edition:         "",
			ProductName:     "Nextcloud mock",
			ExtendedSupport: false,
		}
		respondWithJSON(w, http.StatusOK, response)
		return
	}

	if r.URL.Path == "/index.php/login/v2" {
		response := NextcloudLoginV2{
			Poll: NextcloudPollEndpoint{
				Token:    "abcdefg",
				Endpoint: "http://localhost:8080/index.php/login/v2/poll",
			},
			Login: "http://localhost:8080/index.php/login?token=abcdefg",
		}

		respondWithJSON(w, http.StatusOK, response)
		return
	}

	if r.URL.Path == "/index.php/login/v2/poll" {
		// TODO get bodyBytes
		bodyBytes := ""
		tokenParts := strings.Split(string(bodyBytes), "=")
		if len(tokenParts) != 2 || tokenParts[0] != "token" {
			return
		}

		// TODO: check token authentication
		if tokenParts[1] != "abcdefg" {
			respondWithJSON(w, http.StatusNotFound, []empty{})
			return
		}

		response := NextcloudPollAuth{
			Server:      "http://localhost:8080",
			LoginName:   "admin",
			AppPassword: "admin", // TODO
		}
		respondWithJSON(w, http.StatusOK, response)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/remote.php/dav/") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/remote.php")
		var re = regexp.MustCompile(`files/[^/]+/`)
		r.URL.Path = re.ReplaceAllString(r.URL.Path, "")
		fmt.Println("dav URL", r.URL.Path)
		h.webdavHandler.ServeHTTP(w, r)
		return
	}

	fmt.Println("unknown request, forwarding to webdav")
	h.webdavHandler.ServeHTTP(w, r)
}

func NewResponseLoggingHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// switch out response writer for a recorder
		// for all subsequent handlers
		c := httptest.NewRecorder()
		next(c, r)

		// copy everything from response recorder
		// to actual response writer
		for k, v := range c.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(c.Code)
		c.Body.WriteTo(w)

	}
}

func main() {
	fmt.Println("Starting...")

	st := storage.NewFs("/files")

	webdavHandler := webdav.NewXwebdavHandler(st)

	chi.RegisterMethod("PROPFIND")
	chi.RegisterMethod("PROPPATCH")
	chi.RegisterMethod("MKCOL")
	chi.RegisterMethod("COPY")
	chi.RegisterMethod("MOVE")
	chi.RegisterMethod("LOCK")
	chi.RegisterMethod("UNLOCK")
	r := chi.NewRouter()

	ncHandler := BaseHandler{webdavHandler: webdavHandler}
	r.Mount("/nextcloud", ncHandler)
	r.Mount("/index.php", ncHandler)
	r.Mount("/", ncHandler)
	r.Mount("/dav", webdavHandler)
	err := http.ListenAndServe("localhost:8080", r)
	if err != nil {
		panic(err.Error())
	}
}
