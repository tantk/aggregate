package data

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	config "golive/common/config"
)

//QooNavigate :
func QooNavigate(searchTerm string) string {
	conf := config.GetConfig()
	logging := config.GetInstance(conf.ShopeeLogs)
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(logging.Info.Printf),
	)

	ctx, cancel = context.WithTimeout(ctx, 100*time.Second)
	defer cancel()
	baseSearchURL := "https://www.qoo10.sg/s?keyword="
	searchURL := baseSearchURL + url.QueryEscape(searchTerm)
	chromedp.Run(ctx, chromedp.Navigate(searchURL))
	if err := chromedp.Run(ctx, chromedp.Navigate(searchURL)); err != nil {
		logging.Error.Println("Could not navigate to", searchURL)
		return ""
	}
	var outer string
	//scroll to end and extract search area first
	var res bool
	if err := chromedp.Run(ctx,
		chromedp.EvaluateAsDevTools(`function scrollEnd() {
			curpos = document.documentElement.scrollTop;
			lastValue=0
			while (true) {
				window.scrollBy(0, 150);
				curpos=document.documentElement.scrollTop;
				if(lastValue == curpos)
				{
					break
				}
				lastValue=curpos
				console.log(curpos)
			  }
			return true;
		  };
		  scrollEnd();`, &res),
		chromedp.Sleep(1*time.Second),
		chromedp.OuterHTML("#div_search_area", &outer)); err != nil && res {
		logging.Error.Println("Could not scroll to #div_paging_list and extract #div_search_result_list")
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(outer))
	check(err)
	//Find out number of elements in each of the table
	doc.Find("div.bd_lst_item").Each(func(i0 int, s0 *goquery.Selection) {
		tableL := 0
		id, _ := s0.Attr("id")
		fmt.Println(id)
		s0.Children().Each(func(i1 int, s1 *goquery.Selection) {
			s0.Find(`#div_search_result_list >table>tbody`).Children().Each(func(i2 int, s2 *goquery.Selection) {
				tableL++
			})
		})
		fmt.Println(tableL)
		for u := 1; u <= tableL; u++ {
			sel := fmt.Sprintf(`#%s >table>tbody>tr:nth-child(%d)`, id, u)
			if err := chromedp.Run(ctx,
				chromedp.ScrollIntoView(sel, chromedp.BySearch),
				chromedp.Sleep(200*time.Millisecond),
				chromedp.WaitReady(sel, chromedp.BySearch)); err != nil {
				logging.Error.Println("Error in scroll wait loop number", u, id)
			}
			fmt.Println(sel, u, id)
		}
	})

	if err := chromedp.Run(ctx, chromedp.OuterHTML("#div_search_area", &outer)); err != nil {
		logging.Error.Println("Error saving html", searchURL)
		return ""
	}
	// f, err := os.Create("logs/qoo.xml")
	// check(err)
	// w := bufio.NewWriter(f)
	// n4, err := w.WriteString(outer)
	// check(err)
	// logging.Info.Printf("wrote %d bytes\n", n4)
	// w.Flush()
	return outer
}

//QooSearch :
func QooSearch(searchTerm string, iter int, saveHTML bool) []Item {
	items = make([]Item, 0, 50)
	conf := config.GetConfig()
	logging := config.GetInstance(conf.ShopeeLogs)

	var outer string

	logging.Info.Println("visiting")
	outer = QooNavigate(searchTerm)
	logging.Info.Println("done visiting")
	if saveHTML {
		f, err := os.Create("logs/qoo.xml")
		check(err)
		w := bufio.NewWriter(f)
		n4, err := w.WriteString(outer)
		check(err)
		logging.Info.Printf("wrote %d bytes\n", n4)
		w.Flush()
	}

	//Convert []byte to string and print to screen
	// text, err := ioutil.ReadFile("logs/qoo.xml")
	// outer = string(text)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(outer))
	check(err)

	doc.Find("tbody").Children().Each(func(i int, s *goquery.Selection) {
		fmt.Println(i)
		//For each item found, get the band and title
		nameTag := s.Find(`td.td_item > div > div.sbj > a`)
		var rating float64
		link, _ := nameTag.Attr("href")
		name, _ := nameTag.Attr("title")
		imgTag := s.Find(`td.td_thmb > div > a > img`)
		imgURL, _ := imgTag.Attr("src")
		if strings.Contains(strings.ToLower(imgURL), "loading") {
			imgURL, _ = imgTag.Attr("gd_src")
		}
		priceTag := s.Find(`td.td_prc > div > strong`)
		sprice := priceTag.Text()
		sprice = strings.ReplaceAll(sprice[2:], ",", "")
		price, _ := strconv.ParseFloat(sprice, 64)
		ratingTag := s.Find(`td:nth-child(5) > span > span`)
		srating := ratingTag.Text()
		if len(srating) < 8 {
			rating = 0.00
		} else {
			rating, _ = strconv.ParseFloat(string(srating[8]), 64)
		}

		logging.Info.Println("*****Item details*****")
		logging.Info.Printf("Item %d: item url %s  \n", i, link)
		logging.Info.Printf("Item %d: image url %s \n", i, imgURL)
		logging.Info.Printf("Item %d: item name %s \n", i, name)
		logging.Info.Printf("Item : item price")
		logging.Info.Printf("Item %d: item price %v \n", i, price)
		logging.Info.Printf("Item %d: item rating %v \n", i, rating)
		logging.Info.Println("*****End of Item details*****")

		item := Item{
			ItemURL:         link,
			ImageURL:        imgURL,
			Name:            name,
			PriceRange:      false,
			Price:           price,
			PriceMin:        0,
			PriceMax:        0,
			Rating:          rating,
			Platform:        "Qoo10",
			SearchIteration: iter,
			SearchTerm:      searchTerm,
		}
		items = append(items, item)

	})
	e := BulkInsertItems(items)
	fmt.Println(e)
	return items
}
