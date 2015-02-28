package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type FilterPipeline struct {
	Data   interface{}
	Filter interface{}
}

func setupHandlers(host string) {

	// listT := template.New("list.html")
	// listT, err := listT.ParseFiles(
	listT, err := template.ParseFiles(
		"templates/list.html",
		"templates/navbar.html",
		"templates/benchlist.html",
		"templates/titles.html",
		"templates/scripts.html",
		"templates/stylesheets.html")
	if err != nil || listT == nil {
		log.Fatal(err)
	}
	compareT, err := template.ParseFiles(
		"templates/compare.html",
		"templates/navbar.html",
		"templates/benchlist.html",
		"templates/titles.html",
		"templates/scripts.html",
		"templates/stylesheets.html")
	if err != nil {
		log.Fatal(err)
	}
	indexT, err := template.ParseFiles(
		"templates/index.html",
		"templates/filters.html",
		"templates/index_list.html",
		"templates/index_compare.html",
		"templates/navbar.html",
		"templates/scripts.html",
		"templates/stylesheets.html")
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
		depth, err := strconv.ParseInt(r.FormValue("depth"), 10, 64)
		if err != nil {
			depth = -1
		}

		tables, err := MakeListTables(host, filter, focus, int(depth))
		if err != nil {
			// That's bad? Report it maybe?
			// This means having an error template
		}

		err = listT.Execute(w, struct {
			Filter string
			Focus  string
			Tables []BenchListTable
		}{
			filter,
			focus,
			tables,
		})
		if err != nil {
			log.Println(err)
		}
	})

	http.HandleFunc("/compare", func(w http.ResponseWriter, r *http.Request) {
		filter := r.FormValue("filter")
		focus := r.FormValue("focus")
		spec := r.FormValue("spec")
		values := r.FormValue("values")
		ignore := r.FormValue("ignore")
		depth, err := strconv.ParseInt(r.FormValue("depth"), 10, 64)
		if err != nil {
			depth = -1
		}
		tables, err := MakeCompareTables(host, spec, values, ignore, filter, focus, int(depth))
		if err != nil {
		}

		err = compareT.Execute(w, struct {
			Filter string
			Focus  string
			Spec   string
			Values string
			Ignore string
			Tables []BenchCompareTable
		}{
			filter,
			focus,
			spec,
			values,
			ignore,
			tables,
		})
		if err != nil {
			log.Println(err)
		}
	})

	http.Handle("/static/", http.FileServer(http.Dir(".")))
}
