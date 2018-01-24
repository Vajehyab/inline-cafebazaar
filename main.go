package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Suggestion struct {
	Title  string
	Result string
	Source string
}

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

func doSearch(w http.ResponseWriter, query, deviceID string) {
	user := &User{}
	user.DeviceID = deviceID
	user.Get()

	result, err := sendRequest(query, user.EncodeDictionary())
	if err != nil {
		w.Write(render("error", nil))
		return
	}

	suggestions := make([]Suggestion, 0)

	suggestion, err := getSuggestions(query)
	if err != nil {
		w.Write(render("error", nil))
		return
	}

	for _, v := range suggestion.Data.Suggestion {
		if len(suggestions) > 2 {
			break
		}
		resp, err := sendRequest(v, user.EncodeDictionary())
		if err != nil {
			w.Write(render("error", nil))
			return
		}
		if resp.Data.NumFound > 0 {
			s := Suggestion{}

			if resp.Data.Results[0].Title == query {
				continue
			}

			s.Title = resp.Data.Results[0].Title
			s.Source = resp.Data.Results[0].Source
			s.Result = resp.Data.Results[0].Text

			suggestions = append(suggestions, s)
		}
	}

	dbQuery := &Query{}
	dbQuery.Query = query
	dbQuery.DeviceID = user.DeviceID
	dbQuery.Create()

	user.Save()

	w.Write(render("search", map[string]interface{}{
		"result":     result.Data,
		"status":     true,
		"query":      query,
		"suggestion": suggestions,
	}))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	payload := r.FormValue("payload")
	var Payload struct {
		Word string `json:"word"`
	}
	json.Unmarshal([]byte(payload), &Payload)

	deviceID := r.FormValue("deviceID")

	doSearch(w, Payload.Word, deviceID)
}

func searchHeaderHandler(w http.ResponseWriter, r *http.Request) {
	query := mux.Vars(r)["query"]
	if query == "" {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	deviceID := r.FormValue("deviceID")

	doSearch(w, query, deviceID)
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
	portNumberPtr := flag.Int("port", 7789, "http port")
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

	srv := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf("0.0.0.0:%d", *portNumberPtr),
	}
	log.Println("Running on", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
