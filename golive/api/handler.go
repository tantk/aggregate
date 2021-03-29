package api

import (
	"fmt"
	"golive/common/config"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

//Router create router
func Router() *mux.Router {

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/search", Search)
	router.HandleFunc("/api/v1/latestSearch", LatestSearch)
	return router
}

//StartServer :
func StartServer() {

	//need to use cors to allow using REST api in the same system
	router := Router()
	//allow all domain for local testing
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowCredentials: true,
	})
	conf := config.GetConfig()
	logging := config.GetInstance(conf.ServerLogs)
	handler := c.Handler((router))

	server := &http.Server{
		Addr:         conf.RESTport,
		Handler:      config.Tracing()(logging.Infologging()(handler)),
		ErrorLog:     logging.Error,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	//data.QooSearch("peachy princess", true)
	//data.ShopeeNavigate("ryzen cpu")
	//data.ShopeeSearch("peachy princess", 1, true)
	//items, _ := data.GetAllItems()
	fmt.Println("Listening at port ", conf.RESTport)
	logging.Trace.Fatal(server.ListenAndServe())
}
