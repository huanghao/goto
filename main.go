package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var Mapping map[string]string = make(map[string]string)

func index(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	switch {
	case r.Method == "POST":
		name, url := r.Form["name"][0], r.Form["url"][0]
		Mapping[name] = url
		http.Redirect(w, r, url, 301)
	case r.URL.Path == "/":
		name := r.Form["name"]
		if len(name) > 0 {
			http.Redirect(w, r, "/"+name[0], 301)
		} else {
			t, _ := template.ParseFiles("index.html")
			context := make(map[string]string)
			t.Execute(w, context)
		}
	default:
		name := r.URL.Path[1:]
		url := Mapping[name]
		http.Redirect(w, r, url, 301)
	}
}

func main() {
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
