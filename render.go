package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/oxtoacart/bpool"
)

var bufpool *bpool.BufferPool
var templates *template.Template
var funcMap = make(template.FuncMap)

func init() {
	bufpool = bpool.NewBufferPool(128)
	funcMap["getValue"] = func(m map[string]bool, item string) bool {
		return m[item]
	}
	templates = template.New("").Funcs(funcMap)
	filepath.Walk("view/template", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".xml") {
			log.Println("Reading template", path)
			templates.ParseFiles(path)
		}
		return nil
	})
}

func render(name string, data interface{}) []byte {
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := templates.ExecuteTemplate(buf, name, data)
	if err != nil {
		log.Println("Error on rendering", name, ":", err)
	}

	return buf.Bytes()
}
