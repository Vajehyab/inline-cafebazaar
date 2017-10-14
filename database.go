package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var database *gorm.DB

func init() {
	var err error
	database, err = gorm.Open("sqlite3", "database.db")
	if err != nil {
		panic(err)
	}

	qTable := &query{}

	if !database.HasTable(qTable) {
		database.CreateTable(qTable)
	}

	database.AutoMigrate(qTable)
}

type query struct {
	gorm.Model
	Query    string
	DeviceID string
}

func (q *query) Save() {
	database.Model(q).Save(q)
}

func (q *query) Create() {
	database.Model(q).Create(q)
}

func getQueries(deviceID string, limit int, q string) []query {
	list := []query{}
	database.Model(&query{}).Select("distinct(query)").Where("device_id = ? and query <> ?", deviceID, q).Order("created_at desc").Limit(limit).Scan(&list)
	return list
}
