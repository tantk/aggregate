package data

import (
	"fmt"
	"strings"

	config "golive/common/config"
	//mysql
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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

//Items slice of Item
type Items []Item

//ItemsReturn return format to frontend
type ItemsReturn struct {
	Items Items
}

//Search :
type Search struct {
	ID              int    `json:"id" db:"id"`
	ShopeeTime      int    `json:"ShopeeTime" db:"ShopeeTime"`
	Qoo10Time       int    `json:"Qoo10Time" db:"Qoo10Time"`
	SearchTerm      string `json:"SearchTerm" db:"SearchTerm"`
	SearchDateTime  string `json:"SearchDateTime" db:"SearchDatetime"`
	Status          string `json:"Status" db:"Status"`
	SearchIteration int    `json:"SearchIteration" db:"SearchIteration"`
}

// InSearchQuery json format for accepting search queries
type InSearchQuery struct {
	SearchTerm string `json:"SearchTerm"`
}

//Seller :
type Seller struct {
	Name  string
	Items map[string]string
}

// sqlx package
func dbConn() (db *(sqlx.DB)) {
	conf := config.GetConfig()
	dataSourceName := conf.DbUser + ":" + conf.DbPW + "@tcp(" + conf.DbHost + ":" + conf.DbPort + ")/" + conf.DbName + "?charset=utf8mb4&parseTime=true"
	db, err := sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

//InsertSearch : hard coded table name for simplicity since it is unlikely to change
func InsertSearch(s Search) {
	db := dbConn()
	query := fmt.Sprintf(
		`INSERT INTO search (
			SearchTerm, SearchIteration, ShopeeTime, Qoo10Time, SearchDatetime, Status
			) VALUES ('%s', %d, %d, %d,'%s','%s')`,
		s.SearchTerm, s.SearchIteration, s.ShopeeTime, s.Qoo10Time,
		s.SearchDateTime, s.Status)
	_, err := db.Query(query)
	check(err)
	defer db.Close()
}

//GetSearchBySearchTerm :
func GetSearchBySearchTerm(term string) (Search, bool) {
	search := Search{}
	db := dbConn()
	query := fmt.Sprintf(`SELECT * FROM search where SearchTerm='%s'`, term)
	err := db.Get(&search, query)
	if err != nil {
		fmt.Println(err)
		return search, false
	}
	defer db.Close()
	return search, true
}

//GetMaxIteration :
func GetMaxIteration(term string) (int, bool) {
	var maxIter int
	db := dbConn()
	query := fmt.Sprintf(`SELECT max(SearchIteration) FROM search where SearchTerm ='%s')`, term)
	err := db.Get(&maxIter, query)
	if err != nil {
		fmt.Println(err)
		return 0, false
	}
	defer db.Close()
	return maxIter, true
}

//GetLatestSearchBySearchTerm :
func GetLatestSearchBySearchTerm(term string) (Search, bool) {
	search := Search{}
	db := dbConn()
	query := fmt.Sprintf(`Select * from search where SearchTerm ='%s'and SearchIteration=(SELECT max(SearchIteration) FROM search where SearchTerm ='%s');`, term, term)
	err := db.Get(&search, query)
	if err != nil {
		fmt.Println(err)
		return search, false
	}
	defer db.Close()
	return search, true
}

//GetLatestSearch :
func GetLatestSearch() ([]Search, bool) {
	var searches []Search
	db := dbConn()
	query := fmt.Sprintf(`SELECT *
	FROM search
	INNER JOIN (SELECT max(SearchIteration) as SearchIteration,SearchTerm FROM search group by SearchTerm) as gr
	ON gr.SearchIteration = search.SearchIteration
	and 
	gr.searchTerm = search.searchTerm;`)
	err := db.Select(&searches, query)
	if err != nil {
		fmt.Println(err)
		return searches, false
	}
	defer db.Close()
	return searches, true
}

//GetLatestItemsBySearchTerm :
func GetLatestItemsBySearchTerm(term string) (Items, bool) {
	items := make(Items, 0)
	db := dbConn()
	query := fmt.Sprintf(`Select * from Items where SearchTerm ='%s'and SearchIteration=(SELECT max(SearchIteration) FROM Items where SearchTerm ='%s');`, term, term)
	err := db.Select(&items, query)
	if err != nil {
		fmt.Println(err)
		return items, false
	}
	defer db.Close()
	return items, true
}

//GetAllItems :
func GetAllItems() (Items, bool) {
	items := make(Items, 0)
	db := dbConn()
	query := fmt.Sprintf(`Select * from Items;`)
	err := db.Select(&items, query)
	if err != nil {
		fmt.Println(err)
		return items, false
	}
	defer db.Close()
	return items, true
}

//BulkInsertItems :
func BulkInsertItems(items Items) error {
	db := dbConn()
	valueStrings := make([]string, 0, len(items))
	valueArgs := make([]interface{}, 0, len(items)*14)
	for _, post := range items {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, post.ItemURL)
		valueArgs = append(valueArgs, post.ImageURL)
		valueArgs = append(valueArgs, post.Name)
		valueArgs = append(valueArgs, post.PriceRange)
		valueArgs = append(valueArgs, post.Price)
		valueArgs = append(valueArgs, post.PriceMin)
		valueArgs = append(valueArgs, post.PriceMax)
		valueArgs = append(valueArgs, post.QuantityAvailable)
		valueArgs = append(valueArgs, post.Platform)
		valueArgs = append(valueArgs, post.SellerName)
		valueArgs = append(valueArgs, post.Sales)
		valueArgs = append(valueArgs, post.Rating)
		valueArgs = append(valueArgs, post.SearchTerm)
		valueArgs = append(valueArgs, post.SearchIteration)
	}
	stmt := fmt.Sprintf("INSERT INTO Items (ItemURL, ImageURL, Name, PriceRange, Price, PriceMin, PriceMax, QuantityAvailable, Platform, SellerName,Sales, Rating, SearchTerm, SearchIteration) VALUES %s",
		strings.Join(valueStrings, ","))
	_, err := db.Exec(stmt, valueArgs...)

	defer db.Close()
	return err
}

//UpdateSearch :
func UpdateSearch(search Search) {
	db := dbConn()
	query := fmt.Sprintf(
		`UPDATE search SET ShopeeTime=%d, Qoo10Time=%d, Status='%s' WHERE SearchTerm='%s' and SearchIteration=%d;`,
		search.ShopeeTime, search.Qoo10Time, search.Status, search.SearchTerm, search.SearchIteration)
	_, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
}
