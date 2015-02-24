package main

import "net/http"

func setupHandlers() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/list.html")
	})

	http.HandleFunc("/compare", func(w http.ResponseWriter, r *http.Request) {
	})

	http.Handle("/static/", http.FileServer(http.Dir(".")))
}
