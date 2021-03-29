package controller

import (
	"fmt"
	config "goliveWeb/config"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

//Router create router
func Router() *mux.Router {
	fs := http.FileServer(http.Dir("./static"))
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	router.HandleFunc("/", Index)
	router.HandleFunc("/search", SearchPage)
	router.HandleFunc("/searchReport", SearchReport)
	router.Handle("/favicon.ico", http.NotFoundHandler())
	return router
}

//StartServer :
func StartServer() {

	//need to use cors to allow using REST api in the same system
	router := Router()
	tpl = template.Must(template.ParseGlob("templates/*.html"))
	//allow all domain for local testing
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowCredentials: true,
	})
	conf := config.GetConfig()
	logging := config.GetInstance(conf.WebLogs)
	handler := c.Handler((router))

	server := &http.Server{
		Addr:     conf.WebPort,
		Handler:  config.Tracing()(logging.Infologging()(handler)),
		ErrorLog: logging.Error,
	}

	fmt.Println("Listening at port ", conf.WebPort)
	logging.Trace.Fatal(server.ListenAndServe())
}
