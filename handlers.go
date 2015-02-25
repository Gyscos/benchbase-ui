package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func setupHandlers(host string) {
	listT, err := template.ParseFiles("templates/list.html", "templates/navbar.html", "templates/scripts.html", "templates/stylesheets.html")
	if err != nil {
		log.Fatal(err)
	}
	indexT, err := template.ParseFiles("templates/index.html", "templates/navbar.html", "templates/scripts.html", "templates/stylesheets.html")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := indexT.Execute(w, nil)
		if err != nil {
			log.Println(err)
		}
	})

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		filter := r.FormValue("filter")
		focus := r.FormValue("focus")
		depth, _ := strconv.ParseInt(r.FormValue("depth"), 10, 64)
		err := listT.Execute(w, struct {
			Host   string
			Filter string
			Depth  int64
			Focus  string
		}{
			host,
			filter,
			depth,
			focus,
		})
		if err != nil {
			log.Println(err)
		}
	})

	http.HandleFunc("/compare", func(w http.ResponseWriter, r *http.Request) {
	})

	http.Handle("/static/", http.FileServer(http.Dir(".")))
}
