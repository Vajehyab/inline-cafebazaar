package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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
		"/setting",
		settingHandler,
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

func settingHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := &User{}
	user.DeviceID = r.FormValue("deviceID")
	user.Get()

	if user.ID == 0 {
		user.SetDictionary(user.GetAllDictionaries())
		user.Create()
	}

	m := make(map[string]bool)
	input := r.FormValue("payload")
	if len(input) > 60 {
		input = strings.Replace(input, `"permittedData":{},`, "", -1)
		json.Unmarshal([]byte(input), &m)
		user.SetDictionary(m)
	}

	user.Save() // re-new updated_at field.

	data := make(map[string]interface{})
	data["user"] = user
	data["dictionary"] = user.GetDictionary()
	w.Write(render("setting", data))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	payload := r.FormValue("payload")
	var Payload struct {
		Word string `json:"word"`
	}
	json.Unmarshal([]byte(payload), &Payload)

	user := &User{}
	user.DeviceID = r.FormValue("deviceID")
	user.Get()

	if user.ID == 0 {
		user.SetDictionary(user.GetAllDictionaries())
		user.Create()
	} else {
		user.Save()
	}

	result, err := sendRequest(Payload.Word, user.EncodeDictionary())
	if err != nil {
		w.Write(render("error", nil))
		return
	}

	dbQuery := &Query{}
	dbQuery.Query = Payload.Word
	dbQuery.DeviceID = user.DeviceID
	dbQuery.Save()

	w.Write(render("search", map[string]interface{}{
		"result": result.Data,
		"status": true,
		"query":  Payload.Word,
	}))
}

func searchHeaderHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	query := mux.Vars(r)["query"]
	if query == "" {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	user := &User{}
	user.DeviceID = r.FormValue("deviceID")
	user.Get()

	if user.ID == 0 {
		user.SetDictionary(user.GetAllDictionaries())
		user.Create()
	} else {
		user.Save()
	}

	result, err := sendRequest(query, user.EncodeDictionary())
	if err != nil {
		w.Write(render("error", nil))
		return
	}

	dbQuery := &Query{}
	dbQuery.Query = query
	dbQuery.DeviceID = user.DeviceID
	dbQuery.Save()

	w.Write(render("search", map[string]interface{}{
		"result": result.Data,
		"status": true,
		"query":  query,
	}))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	user := &User{}
	user.DeviceID = r.FormValue("deviceID")
	user.Get()

	if user.ID == 0 {
		user.SetDictionary(user.GetAllDictionaries())
		user.Create()
	} else {
		user.Save()
	}

	data := make(map[string]interface{})
	data["user"] = user
	w.Write(render("index", data))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(render("index", nil))
}

func main() {
	logFilePtr := flag.String("log", "requests.log", "log file for requests")
	portNumberPtr := flag.Int("port", 9989, "http port")
	flag.Parse()

	log.Println("Starting ...")

	router := mux.NewRouter().StrictSlash(false)
	for _, route := range routes {
		router.
			Methods("GET", "POST").
			Path(route.Pattern).
			Handler(route.HandlerFunc)
	}

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	file, err := os.OpenFile(*logFilePtr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	defer file.Close()

	srv := &http.Server{
		Handler: handlers.LoggingHandler(file, router),
		Addr:    fmt.Sprintf("0.0.0.0:%d", *portNumberPtr),
	}
	log.Println("Running on", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
