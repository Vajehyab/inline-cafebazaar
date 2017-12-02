package main

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	database   *gorm.DB
	queryModel = new(Query)
	userModel  = new(User)
)

func init() {
	var err error
	database, err = gorm.Open("sqlite3", "database.db")
	if err != nil {
		panic(err)
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
	for key, value := range m {
		if value {
			output += key + ","
		}
	}
	output = output[:len(output)-1]

	return output
}

func (u *User) Save() {
	database.Model(u).Save(u)
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
	bytes, _ := json.Marshal(input)
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
