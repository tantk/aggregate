package data

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	config "golive/common/config"
)

var items []Item

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//ShopeeNavigate : navigator that retrieve data from website
func ShopeeNavigate(searchTerm string) string {
	conf := config.GetConfig()
	logging := config.GetInstance(conf.ShopeeLogs)
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(logging.Info.Printf),
	)
	ctx, cancel = context.WithTimeout(ctx, 100*time.Second)
	defer cancel()
	baseSearchURL := "https://shopee.sg/search?keyword="
	searchURL := baseSearchURL + url.QueryEscape(searchTerm)
	fmt.Println(searchURL)
	var outer string
	chromedp.Run(ctx, chromedp.Navigate(searchURL), chromedp.Sleep(1*time.Second))

	for u := 1; u <= 50; u++ {
		sel := fmt.Sprintf(`#main > div > div > div.container> div> div.shopee-search-item-result > div.row.shopee-search-item-result__items > div:nth-child(%d) > div > a > div > div > img`, u)
		fmt.Println(sel)
		if err := chromedp.Run(ctx,
			chromedp.ScrollIntoView(sel, chromedp.BySearch),
			chromedp.WaitReady(sel, chromedp.BySearch),
			chromedp.WaitVisible(sel, chromedp.BySearch)); err != nil {
			logging.Error.Println("Error in scroll wait loop number", u)
		}
	}
	chromedp.Run(ctx, chromedp.OuterHTML("div.shopee-search-item-result", &outer))
	// f, err := os.Create("logs/shopeetest.xml")
	// check(err)
	// w := bufio.NewWriter(f)
	// n4, err := w.WriteString(outer)
	// check(err)
	// logging.Info.Printf("wrote %d bytes\n", n4)
	// w.Flush()
	return outer
}

//ShopeeSearch : calls ShopeeNavigate and parses the xml into structs
func ShopeeSearch(searchTerm string, iter int, saveHTML bool) []Item {
	var outer string
	items = make([]Item, 0, 50)
	conf := config.GetConfig()
	logging := config.GetInstance(conf.ShopeeLogs)

	logging.Info.Println("visiting")
	outer = ShopeeNavigate(searchTerm)
	logging.Info.Println("done visiting")
	if saveHTML {
		f, err := os.Create("logs/shopee.xml")
		check(err)
		w := bufio.NewWriter(f)
		n4, err := w.WriteString(outer)
		check(err)
		logging.Info.Printf("wrote %d bytes\n", n4)
		w.Flush()
	}
	baseURL := "https://shopee.sg"
	// Convert []byte to string and print to screen
	//outer := ShopeeNavigate(searchTerm)
	// text, err := ioutil.ReadFile("logs/shopee.xml")
	// outer = string(text)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(outer))
	check(err)
	reFlts := regexp.MustCompile(`\d+\.\d+|\d+`)

	doc.Find(".shopee-search-item-result__items").Children().Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		linkTag := s.Find(`div a[data-sqe="link"]`)
		link, _ := linkTag.Attr("href")
		imageTag := s.Find(`div a[data-sqe="link"] div div img`)
		imageURL, _ := imageTag.Attr("src")
		nameTag := s.Find(`div[data-sqe="name"]`).Children().Eq(0)
		name := nameTag.Text()
		priceTag := s.Find(`div > a > div > div > div > div > span`)
		tprice := priceTag.Text()
		var ratings float64
		ratings = 0.00
		s.Find(`div > a > div > div > div > div > div > div`).Children().Each(func(star int, rating *goquery.Selection) {
			//fmt.Printf("Star %d  \n", star)
			ratingtag := rating.Find(`.shopee-rating-stars__lit`)
			rate, _ := ratingtag.Attr("style")
			//fmt.Printf("Item %d: item rating stars %s  \n", i, rate)
			//reFlts.FindAllSubmatch([]byte(rate), -1)
			curStar := string(reFlts.Find([]byte(rate)))

			if fRating, err := strconv.ParseFloat(curStar, 32); err == nil {
				fRating = math.Round(fRating*100) / 100
				ratings += fRating
			}

		})
		ratings = ratings / 100
		logging.Info.Println("*****Item details*****")
		logging.Info.Printf("Item %d: item url %s  \n", i, baseURL+link)
		logging.Info.Printf("Item %d: image url %s \n", i, imageURL)
		logging.Info.Printf("Item %d: item name %s \n", i, name)
		logging.Info.Printf("Item : item price")
		// re := regexp.MustCompile(`\$\d+(?:.(\d+))?`)
		//fmt.Printf("%q\n", re.FindAllSubmatch([]byte(price), -1))
		split := strings.Split(tprice, "$")
		var prices []float64
		prices = make([]float64, 0)
		for _, v := range split {
			//fmt.Println("splitted", v)
			v = strings.ReplaceAll(v, ",", "")
			if fPrice, err := strconv.ParseFloat(v, 64); err == nil {
				fPrice = math.Round(fPrice*100) / 100
				prices = append(prices, fPrice)
			}
		}

		sort.Float64s(prices)
		logging.Info.Println(prices)
		logging.Info.Printf("Item %d: item rating %v \n", i, ratings)

		var price, priceMin, priceMax float64
		priceRange := false
		if len(prices) > 1 {
			priceRange = true
		}
		if priceRange {
			priceMin = math.Min(prices[0], prices[1])
			priceMax = math.Max(prices[0], prices[1])
			price = (priceMax + priceMin) / 2
		} else {
			price = prices[0]
		}
		item := Item{
			ItemURL:         baseURL + link,
			ImageURL:        imageURL,
			Name:            name,
			PriceRange:      priceRange,
			Price:           price,
			PriceMin:        priceMin,
			PriceMax:        priceMax,
			Rating:          ratings,
			Platform:        "Shopee",
			SearchIteration: iter,
			SearchTerm:      searchTerm,
		}
		//fmt.Println(priceRange, price)
		items = append(items, item)
		logging.Info.Println("*****End of Item details*****")
	})

	return items
}
