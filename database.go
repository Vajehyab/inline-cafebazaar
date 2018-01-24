package main

import (
	"encoding/json"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	database   *gorm.DB
	queryModel = new(Query)
	userModel  = new(User)

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
)

func init() {
	var err error
	database, err = gorm.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatalln(err)
	}

	if !database.HasTable(queryModel) {
		database.CreateTable(queryModel)
	}
	database.AutoMigrate(queryModel)

	if !database.HasTable(userModel) {
		database.CreateTable(userModel)
	}
	database.AutoMigrate(userModel)
}

type User struct {
	gorm.Model
	DeviceID   string
	Dictionary string // json
}

func (u *User) Get() {
	database.Model(u).Find(u, "device_id = ?", u.DeviceID)
}

func (u *User) EncodeDictionary() string {
	output := ""

	m := u.GetDictionary()
	for _, key := range keys {
		if m[key] {
			output += key + ","
		}
	}
	output = output[:len(output)-1]

	return output
}

func (u *User) Save() {
	database.Model(u).Save(u)
}

func (u *User) GetAllDictionaries() map[string]bool {
	m := make(map[string]bool)

	for _, i := range keys {
		m[i] = true
	}

	return m
}

func (u *User) CheckEmptyDictionary() bool {
	if u.Dictionary == "" {
		m := u.GetAllDictionaries()
		u.SetDictionary(m)
		return true
	}
	return false
}

func (u *User) Create() {
	database.Model(u).Create(u)
}

func (u *User) GetDictionary() map[string]bool {
	output := make(map[string]bool)
	json.Unmarshal([]byte(u.Dictionary), &output)
	return output
}

func (u *User) SetDictionary(input map[string]bool) {
	bytes, err := json.Marshal(input)
	if err != nil {
		log.Println("Error on set dictionary", err)
	}
	u.Dictionary = string(bytes)
}

type Query struct {
	gorm.Model
	Query    string
	DeviceID string
}

func (q *Query) Save() {
	database.Model(q).Save(q)
}

func (q *Query) Create() {
	database.Model(q).Create(q)
}

func getQueries(deviceID string, limit int, q string) []Query {
	list := []Query{}
	database.Model(&Query{}).Select("distinct(query)").Where("device_id = ? and query <> ?", deviceID, q).Order("created_at").Limit(limit).Scan(&list)
	return list
}
