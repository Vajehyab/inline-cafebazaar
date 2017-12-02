package main

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/oxtoacart/bpool"
)

var bufpool *bpool.BufferPool
var templates *template.Template
var funcMap = make(template.FuncMap)

func init() {
	bufpool = bpool.NewBufferPool(128)
	templates = template.New("").Funcs(funcMap)
	filepath.Walk("view/template", func(path string, info os.FileInfo, err error) error {
		templates.ParseFiles(path)
		return nil
	})
}

func render(name string, data interface{}) []byte {
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	templates.ExecuteTemplate(buf, name, data)

	return buf.Bytes()
}
