package api

import (
	"encoding/json"
	"fmt"
	"golive/common/config"
	"golive/data"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

func santizeString(s string) string {
	p := bluemonday.UGCPolicy()
	return p.Sanitize(s)
}

//ScrapData go routine , there might be situation where one platform fails and search is updated as complete. i am prioritising data collection over user friendliness
func ScrapData(searchTerm string, iter int) {

	ctimeShopee := make(chan int)
	ctimeQoo10 := make(chan int)

	go func() {
		start1 := time.Now()
		// some computation
		items1 := data.ShopeeSearch(searchTerm, iter, true)
		data.BulkInsertItems(items1)
		ctimeShopee <- int(time.Since(start1)) / 1000
	}()
	go func() {
		start2 := time.Now()
		// some computation
		items2 := data.QooSearch(searchTerm, iter, true)
		data.BulkInsertItems(items2)
		ctimeQoo10 <- int(time.Since(start2)) / 1000
	}()

	timeShopee := <-ctimeShopee
	timeQoo10 := <-ctimeQoo10

	search := data.Search{
		ShopeeTime:      timeShopee,
		Qoo10Time:       timeQoo10,
		SearchTerm:      searchTerm,
		Status:          "Completed",
		SearchIteration: iter}
	data.UpdateSearch(search)
}

//Search :
func Search(w http.ResponseWriter, r *http.Request) {
	conf := config.GetConfig()
	logging := config.GetInstance(conf.ServerLogs)
	switch r.Method {
	case "POST":
		if r.Header.Get("Content-type") == "application/json" {
			var reqC data.InSearchQuery
			reqBody, err := ioutil.ReadAll(r.Body)
			if err == nil {
				json.Unmarshal(reqBody, &reqC)
			} else {
				logging.Error.Println("Error reading body from ", r.UserAgent(), r.Body)
			}
			reqC.SearchTerm = strings.ToLower(santizeString(reqC.SearchTerm))
			if len(reqC.SearchTerm) > 0 {
				var resdata data.ItemsReturn
				now := time.Now()
				latestSearch, ok := data.GetLatestSearchBySearchTerm(reqC.SearchTerm)
				layout := "2006-01-02T15:04:05Z"
				str := latestSearch.SearchDateTime
				t, _ := time.Parse(layout, str)
				expireTime := t.Add(1 * time.Hour)
				expired := expireTime.Before(now)
				fmt.Println(str, t, expireTime)
				//search exist, check for time expiry
				fmt.Println(ok, !expired, latestSearch.Status)
				if ok && !expired && latestSearch.Status != "queue" {

					items, _ := data.GetLatestItemsBySearchTerm(reqC.SearchTerm)
					resdata.Items = items
					json.NewEncoder(w).Encode(resdata)
				} else if latestSearch.Status != "queue" {
					search := data.Search{
						ShopeeTime:      0,
						Qoo10Time:       0,
						SearchTerm:      reqC.SearchTerm,
						SearchDateTime:  time.Now().Format("2006-01-02 15:04:05"),
						Status:          "queue",
						SearchIteration: latestSearch.SearchIteration + 1,
					}
					data.InsertSearch(search)
					go ScrapData(reqC.SearchTerm, latestSearch.SearchIteration+1)
					fmt.Println("not")
					json.NewEncoder(w).Encode(resdata)
				} else {
					//empty response
					json.NewEncoder(w).Encode(resdata)
				}

			} else {
				logging.Error.Println("Error with query content from", r.UserAgent(), reqC.SearchTerm)
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			logging.Error.Println("No json body from", r.UserAgent())
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

//LatestSearch :
func LatestSearch(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		searches, _ := data.GetLatestSearch()

		json.NewEncoder(w).Encode(searches)
	}
}
