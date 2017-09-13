package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	LOG_FILE    = "requests.log"
	PORT_NUMBER = 7766
)

type Route struct {
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"/",
		indexHandler,
	},
	Route{
		"/help/aboutus",
		aboutHandler,
	},
	Route{
		"/help/more",
		helpHandler,
	},
	Route{
		"/index/search/",
		searchHandler,
	},
	Route{
		"/search/{query}",
		searchHeaderHandler,
	},
}

func helpHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(render("more", nil))
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(render("about", nil))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	payload := r.FormValue("payload")
	var Payload struct {
		Word string `json:"word"`
	}
	json.Unmarshal([]byte(payload), &Payload)
	result, err := sendRequest(Payload.Word)
	if err != nil {
		w.Write(render("error", nil))
		return
	}
	w.Write(render("search", map[string]interface{}{
		"result": result.Data,
		"status": true,
	}))
}

func searchHeaderHandler(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)["query"]

	result, err := sendRequest(query)
	if err != nil {
		w.Write(render("error", nil))
		return
	}

	w.Write(render("search", map[string]interface{}{
		"result": result.Data,
		"status": true,
	}))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(render("index", nil))
}

func main() {
	router := mux.NewRouter().StrictSlash(false)
	for _, route := range routes {
		router.
			Methods("POST").
			Path(route.Pattern).
			Handler(route.HandlerFunc)
	}

	file, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	defer file.Close()

	srv := &http.Server{
		Handler: handlers.LoggingHandler(file, router),
		Addr:    fmt.Sprintf("0.0.0.0:%d", PORT_NUMBER),
	}
	log.Fatal(srv.ListenAndServe())
}
