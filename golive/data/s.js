//some javascripts and commented codes. for backup purpose

async function myFunction() {
    function sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
     }
    curpos = document.documentElement.scrollTop;
    lastValue=0
    while (true) {
        window.scrollBy(0, 150);
        await sleep(300)
        curpos=document.documentElement.scrollTop;
        if(lastValue == curpos)
        {
            break
        }
        lastValue=curpos
        console.log(curpos)
      }
    return true;
  }

  function scrollDown() {
    curpos = document.documentElement.scrollTop;
    window.scrollBy(0, 150);
    return curpos;
  }


  	//scroll to end and extract search area first
	//var res int
	// var res bool
	// if err := chromedp.Run(ctx,
	// 	chromedp.EvaluateAsDevTools(`function myFunction() {
	// 		curpos = document.documentElement.scrollTop;
	// 		lastValue=0
	// 		while (true) {
	// 			window.scrollBy(0, 150);
	// 			curpos=document.documentElement.scrollTop;
	// 			if(lastValue == curpos)
	// 			{
	// 				break
	// 			}
	// 			lastValue=curpos
	// 			console.log(curpos)
	// 			}
	// 		return true;
	// 		};
	// 		myFunction();`, &res),
	// 	chromedp.Sleep(1*time.Second),
	// 	chromedp.OuterHTML("div.shopee-search-item-result", &outer)); err != nil && res {
	// 	logging.Error.Println("Could not find div.shopee-search-item-result")
	// 	return ""

	// }
	// lastValue := 1
	// for res != lastValue {
	// 	if err := chromedp.Run(ctx,
	// 		chromedp.EvaluateAsDevTools(`function scrollDown() {
	// 			curpos = document.documentElement.scrollTop;
	// 			window.scrollBy(0, 150);
	// 			return curpos;
	// 		  };
	// 		  scrollDown();`, &res),
	// 		chromedp.Sleep(500*time.Millisecond)); err != nil {
	// 		logging.Error.Println("Error with scrolling")
	// 		return ""
	// 	}
	// 	lastValue = res
	// }

	//doc, err := goquery.NewDocumentFromReader(strings.NewReader(outer))
	//check(err)
	//Find out number of elements in each of the table
	//itemNo := 0


// 	ALTER DATABASE aggre CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

// ALTER TABLE aggre.Items CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
// ALTER TABLE aggre.Items MODIFY ItemURL TEXT CHARSET utf8mb4;
// ALTER TABLE aggre.Items MODIFY Name TEXT CHARSET utf8mb4;


// //ScrapData go routine
// func ScrapData(searchTerm string, iter int) {
// 	start1 := time.Now()
// 	// some computation
// 	items1 := data.ShopeeSearch(searchTerm, iter, true)
// 	shtime := int(time.Since(start1)) / 1000
// 	data.BulkInsertItems(items1)
// 	start2 := time.Now()
// 	items2 := data.QooSearch(searchTerm, iter, true)
// 	qotime := int(time.Since(start2)) / 1000
// 	data.BulkInsertItems(items2)
// 	search := data.Search{
// 		ShopeeTime:      shtime,
// 		Qoo10Time:       qotime,
// 		SearchTerm:      searchTerm,
// 		Status:          "Completed",
// 		SearchIteration: iter}
// 	data.UpdateSearch(search)
// }