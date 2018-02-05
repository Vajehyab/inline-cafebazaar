package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kataras/iris"
)

const (
	DATABASE_FILE = "database.bolt"
	LISTEN_PORT   = 8080
)

type (
	Search struct {
		Context  iris.Context
		Query    string
		DeviceID string
	}
	Suggestion struct {
		Title  string
		Result string
		Source string
	}
)

var (
	keys = []string{
		"sareh",
		"ganjvajeh",
		"slang",
		"fa2en",
		"en2fa",
		"ar2fa",
		"dezfuli",
		"farhangestan",
		"thesis",
		"fa2ar",
		"isfahani",
		"tehrani",
		"wiki",
		"motaradef",
		"quran",
		"bakhtiari",
		"moein",
		"name",
		"gonabadi",
		"mazani",
		"amid",
	}

	db = new(bolt.DB)
)

func main() {
	var err error
	db, err = bolt.Open(DATABASE_FILE, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := iris.New()
	app.RegisterView(iris.HTML("./view/template", ".xml"))

	app.HandleMany("GET POST", "/", indexHandler)
	app.HandleMany("GET POST", "/setting", settingHandler)
	app.HandleMany("GET POST", "/index/search", searchHandler)
	app.HandleMany("GET POST", "/search/{query}", searchHeaderHandler)
	app.HandleMany("GET POST", "/error/", errorHandler)

	app.Run(iris.Addr(fmt.Sprintf(":%d", LISTEN_PORT)), iris.WithoutServerError(iris.ErrServerClosed))
}

func doSearch(ctx iris.Context, deviceID, query string) {
	dictionary := map[string]bool{}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(toBytes(deviceID))

		err := bucket.Put(toBytes("last_used"), toBytes(time.Now()))
		if err != nil {
			return err
		}

		dic := bucket.Get(toBytes("settings"))
		if err = gobDecoder(dic, &dictionary); err != nil {
			return fmt.Errorf("gob error %s", err.Error())
		}

		return nil
	})
	if err != nil {
		ctx.Redirect("/error/?err=" + err.Error())
		return
	}

	encodedDictionary := encodeDictionary(dictionary)

	result, err := getWord(query, encodedDictionary)
	if err != nil {
		ctx.Redirect("/error?err=" + err.Error())
		return
	}

	suggestions := make([]Suggestion, 0)

	suggestion, err := getSuggestion(query)
	if err != nil {
		ctx.Redirect("/error?err=" + err.Error())
		return
	}

	for _, v := range suggestion.Data.Suggestion {
		if len(suggestions) > 2 {
			break
		}
		resp, err := getWord(v, encodedDictionary)
		if err != nil {
			ctx.Redirect("/error?err=" + err.Error())
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

	ctx.ViewData("result", result.Data)
	ctx.ViewData("status", true)
	ctx.ViewData("query", query)
	ctx.ViewData("suggestion", suggestions)
	ctx.View("search.xml")
}

func errorHandler(ctx iris.Context) {
	log.Println(ctx.FormValue("err"))
	ctx.View("error.xml")
}

func searchHeaderHandler(ctx iris.Context) {
	query := ctx.Params().Get("query")
	deviceID := ctx.PostValue("deviceID")
	if deviceID == "" {
		// for GET method, we do not have PostValue.
		deviceID = "unknown_user"
	}
	doSearch(ctx, deviceID, query)
}

func searchHandler(ctx iris.Context) {
	var payload struct {
		Word string `json:"word"`
	}
	json.Unmarshal([]byte(ctx.PostValue("payload")), &payload)

	deviceID := ctx.PostValue("deviceID")
	if deviceID == "" {
		// for GET method, we do not have PostValue.
		deviceID = "unknown_user"
	}

	doSearch(ctx, deviceID, payload.Word)
}

func settingHandler(ctx iris.Context) {
	deviceID := ctx.PostValue("deviceID")
	if deviceID == "" {
		// for GET method, we do not have PostValue.
		deviceID = "unknown_user"
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(toBytes(deviceID))

		err := bucket.Put(toBytes("last_used"), toBytes(time.Now()))
		if err != nil {
			return err
		}

		input := ctx.PostValue("payload")
		if len(input) > 60 {
			// payload is not for empty POST data.
			input = strings.Replace(input, `"permittedData":{},`, "", -1)
			bucket.Put(toBytes("settings"), toBytes(input))

			dictionary := map[string]interface{}{}
			if gobDecoder(toBytes(input), &dictionary) != nil {
				return fmt.Errorf("gob error %s", err.Error())
			}
			ctx.ViewData("dictionary", dictionary)
		}
		return nil
	})
	if err != nil {
		ctx.Redirect("/error/?err=" + err.Error())
		return
	}
	ctx.View("setting.xml")
}

func indexHandler(ctx iris.Context) {
	deviceID := ctx.PostValue("deviceID")
	if deviceID == "" {
		// for GET method, we do not have PostValue.
		deviceID = "unknown_user"
	}

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(toBytes(deviceID))
		if err != nil {
			return err
		}

		err = bucket.Put(toBytes("last_used"), toBytes(time.Now()))
		if err != nil {
			return err
		}

		dictionaries := make(map[string]bool)
		for _, key := range keys {
			dictionaries[key] = true
		}

		err = bucket.Put(toBytes("settings"), toBytes(dictionaries))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		ctx.Redirect("/error/?err=" + err.Error())
		return
	}
	ctx.View("index.xml")
}

func toBytes(a interface{}) []byte {
	switch value := a.(type) {
	case string:
		return []byte(value)
	case int64:
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, value)
		return buf[:n]
	default:
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		enc.Encode(a)
		return buf.Bytes()
	}
}

func gobDecoder(a []byte, output interface{}) error {
	buf := new(bytes.Buffer)
	buf.Write(a)
	enc := gob.NewDecoder(buf)
	return enc.Decode(output)
}
