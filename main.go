package main

import (
	"flag"
	"html/template"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	port = flag.String("p", "8000", "Port number (default 8000)")
)

var stories []story

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)

}
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}

	t, err := template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html", "templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Render the template
	err = t.ExecuteTemplate(w, "base", map[string]interface{}{"Stories": stories})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	flag.Parse()

	go func() {
		for {
			log.Info("======= Fetching new stories =========")
			stories = fetch()
			time.Sleep(20 * time.Second)
		}
	}()
	http.HandleFunc("/", HomeHandler)
	log.Println("Running on localhost:" + *port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+*port, nil))

}
