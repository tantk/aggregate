package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goliveWeb/config"
	"html/template"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

//Item :
type Item struct {
	ID                int     `json:"id" db:"id"`
	ItemURL           string  `json:"ItemURL" db:"ItemURL"`
	ImageURL          string  `json:"ImageURL" db:"ImageURL"`
	Name              string  `json:"Name" db:"Name"`
	PriceRange        bool    `json:"PriceRange" db:"PriceRange"`
	Price             float64 `json:"Price" db:"Price"`
	PriceMin          float64 `json:"PriceMin" db:"PriceMin"`
	PriceMax          float64 `json:"PriceMax" db:"PriceMax"`
	QuantityAvailable int32   `json:"QuantityAvailable" db:"QuantityAvailable"`
	Platform          string  `json:"Platform" db:"Platform"`
	SellerName        string  `json:"SellerName" db:"SellerName"`
	Sales             int32   `json:"Sales" db:"Sales"`
	Rating            float64 `json:"Rating" db:"Rating"`
	SearchTerm        string  `json:"SearchTerm" db:"SearchTerm"`
	SearchIteration   int     `json:"SearchIteration" db:"SearchIteration"`
}

//Search :
type Search struct {
	ID              int     `json:"id" db:"id"`
	ShopeeTime      float64 `json:"ShopeeTime" db:"ShopeeTime"`
	fShopeeTime     float64
	Qoo10Time       float64 `json:"Qoo10Time" db:"Qoo10Time"`
	fQoo10Time      float64
	SearchTerm      string `json:"SearchTerm" db:"SearchTerm"`
	SearchDateTime  string `json:"SearchDateTime" db:"SearchDatetime"`
	Status          string `json:"Status" db:"Status"`
	SearchIteration int    `json:"SearchIteration" db:"SearchIteration"`
}

//Items :
type Items struct {
	Items []Item `json:"Items"`
}

var tpl *template.Template

func santizeString(s string) string {
	p := bluemonday.UGCPolicy()
	return p.Sanitize(s)
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}

//Index : landing page
func Index(res http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(res, "index.html", nil)
}

//SearchPage : search page
func SearchPage(res http.ResponseWriter, req *http.Request) {
	type resData struct {
		Items      Items
		searchTerm string
	}
	resp := resData{}
	searchTerm, ok := req.URL.Query()["SearchTerm"]
	if !ok || len(searchTerm[0]) < 1 {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	search := santizeString(searchTerm[0])
	m := map[string]string{"SearchTerm": search}
	jsonValue, _ := json.Marshal(m)
	conf := config.GetConfig()
	var netClient = &http.Client{
		Timeout: time.Second * 30,
	}

	response, err := netClient.Post(conf.RestURL+"search",
		"application/json", bytes.NewBuffer(jsonValue))

	check(err)
	defer response.Body.Close()
	fmt.Println(response)
	body, err := ioutil.ReadAll(response.Body)
	check(err)
	items := Items{}

	jsonErr := json.Unmarshal(body, &items)
	check(jsonErr)
	resp.Items = items
	resp.searchTerm = search
	if len(items.Items) == 0 {
		http.Redirect(res, req, "/searchReport", http.StatusSeeOther)
	}
	tpl.ExecuteTemplate(res, "search.html", resp)
}

//SearchReport : Search progress
func SearchReport(res http.ResponseWriter, req *http.Request) {
	conf := config.GetConfig()
	var netClient = &http.Client{
		Timeout: time.Second * 30,
	}
	response, err := netClient.Get(conf.RestURL + "latestSearch")
	check(err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	check(err)
	var searches []Search
	jsonErr := json.Unmarshal(body, &searches)
	check(jsonErr)

	for i := range searches {
		searches[i].Qoo10Time = float64(searches[i].Qoo10Time) / 1000000
		searches[i].Qoo10Time = math.Round(searches[i].Qoo10Time*100) / 100
		searches[i].ShopeeTime = float64(searches[i].ShopeeTime) / 1000000
		searches[i].ShopeeTime = math.Round(searches[i].ShopeeTime*100) / 100
	}
	tpl.ExecuteTemplate(res, "searchReport.html", searches)
}
