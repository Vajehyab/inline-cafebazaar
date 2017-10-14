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
	Route{
		"/history/{query}",
		searchHistoryHandler,
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

	if len(Payload.Word) > 0 {
		addToDatabase(r.FormValue("deviceID"), Payload.Word)
	}

	w.Write(render("search", map[string]interface{}{
		"result": result.Data,
		"status": true,
		"list":   getList(r, Payload.Word),
	}))
}

func searchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)["query"]

	result, err := sendRequest(query)
	if err != nil {
		w.Write(render("error", nil))
		return
	}

	w.Write(render("search", map[string]interface{}{
		"result": result.Data,
		"status": true,
		"list":   getList(r, query),
	}))
}

func searchHeaderHandler(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)["query"]

	result, err := sendRequest(query)
	if err != nil {
		w.Write(render("error", nil))
		return
	}

	if len(query) > 0 {
		r.ParseForm()
		addToDatabase(r.FormValue("deviceID"), query)
	}

	w.Write(render("search", map[string]interface{}{
		"result": result.Data,
		"status": true,
		"list":   getList(r, query),
	}))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(render("index", nil))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(render("index", nil))
}

func main() {
	log.Println("Starting ...")

	router := mux.NewRouter().StrictSlash(false)
	for _, route := range routes {
		router.
			Methods("POST").
			Path(route.Pattern).
			Handler(route.HandlerFunc)
	}

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

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

func addToDatabase(deviceID string, q string) {
	tmp := &query{
		DeviceID: deviceID,
		Query:    q,
	}
	tmp.Save()
}

func getList(r *http.Request, q string) []query {
	r.ParseForm()
	deviceID := r.FormValue("deviceID")
	return getQueries(deviceID, 2, q)
}
