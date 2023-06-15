package crawl

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gosimple/slug"
)

type Product struct {
	name           string
	slug           string
	description    string
	assets         string
	facets         string
	optionGroups   string
	optionValues   string
	sku            string
	price          string
	taxCategory    string
	stockOnHand    int8
	trackInventory bool
	variantAssets  string
	variantFacets  string
}
type Translate struct {
	source string
	target string
}

func (p Product) GetHeader() []string {
	var result []string = []string{}
	a := &Product{}
	val := reflect.ValueOf(a).Elem()
	for i := 0; i < val.NumField(); i++ {
		result = append(result, val.Type().Field(i).Name)
	}
	return result
}
func (p Product) ToSlice() []string {
	var result []string = []string{
		fmt.Sprintf(`%s`, p.name),
		fmt.Sprintf(`%s`, p.slug),
		fmt.Sprintf(`%s`, p.description),
		fmt.Sprintf(`%s`, p.assets),
		fmt.Sprintf(`%s`, p.facets),
		fmt.Sprintf(`%s`, p.optionGroups),
		fmt.Sprintf(`%s`, p.optionValues),
		fmt.Sprintf(`%s`, p.sku),
		fmt.Sprintf(`%s`, p.price),
		fmt.Sprintf(`%s`, p.taxCategory),
		fmt.Sprintf(`%d`, p.stockOnHand),
		fmt.Sprintf(``),
		fmt.Sprintf(`%s`, p.variantAssets),
		fmt.Sprintf(`%s`, p.variantFacets),
	}

	return result
}

func translate(source string) string {
	translation := []Translate{
		{source: "Sánh Điệu", target: "Sành Điệu"},
		{source: "Tính Tế", target: "Tinh Tế"},
		{source: "Gợi Cả", target: "Gợi Cảm"},
		{source: "Ngọt Gào", target: "Ngọt Ngào"},
		{source: "Hiện Đai", target: "Hiện DDaij"},
		{source: "Hấp Dấn", target: "Hấp Dẫn"},
		{source: "Ám Áp", target: "Ấm áp"},
		{source: "Ấp Áp", target: "Ấm áp"},
		{source: "Nam Tinh", target: "Nam Tính"},
		{source: "Nam Tinh", target: "Nam Tính"},
		{source: "Rẻ Trung", target: "Trẻ trung"},
		{source: "Khêu Gợi", target: "Khiêu Gợi"},
	}

	for _, s := range translation {
		if source == s.source {
			return s.target
		}
	}

	return source
}

func handleCrawl(info CrawlInfo, w *csv.Writer) {
	c := colly.NewCollector(colly.MaxDepth(2))

	if w == nil {
		file, err := os.Create(info.filename)
		defer file.Close()
		if err != nil {
			log.Fatalln("failed to open file", err)
		}
		w = csv.NewWriter(file)
		headers := Product{}.GetHeader()
		w.Write(headers)
	}

	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL)
	// })
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*thegioinuochoa.*",
		Parallelism: 20,
		RandomDelay: 2 * time.Second,
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	// c.OnResponse(func(r *colly.Response) {
	// 	// fmt.Println("Visited", r.Request.URL)
	// })
	// On every a element which has href attribute call callback
	// regex := regexp.MustCompile(`[^0-9a-zA-Z:,]+`)
	c.OnHTML("#ap_grid-data .product-item", func(e *colly.HTMLElement) {
		detailLink, _ := e.DOM.Children().Find(".product-name").Attr("href")

		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, detailLink)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(detailLink))

	})
	specialRegex := regexp.MustCompile(`[^0-9a-zA-Z]+`)

	c.OnHTML(".product-detail", func(e *colly.HTMLElement) {
		defer w.Flush()
		var product = new(Product)
		product.optionGroups = "size"
		product.taxCategory = "standard"
		product.stockOnHand = 100
		product.trackInventory = false
		product.variantAssets = ""
		product.variantFacets = ""
		product.name = strings.TrimSpace(e.DOM.Find(".product-detail-info #product-title").Text())
		product.slug = slug.Make(product.name)

		var assets []string

		e.ForEach(".product-thumbnail-slider-wrapper .product-thumbnail-slider .product-syn-slider-2-wrapper .gallery .item", func(i int, h *colly.HTMLElement) {
			image, _ := h.DOM.Attr("href")
			assets = append(assets, image)
		})

		if len(assets) == 0 {
			thumb, _ := e.DOM.Find(".product-thumbnail-slider-wrapper .product-thumbnail-slider .product-syn-slider-1-wrapper .item img").Attr("src")
			assets = append(assets, thumb)
		}
		product.assets = strings.Join(assets, "|")

		var attributes []string = []string{info.facets}
		e.ForEach(".product-detail-info .product .product-body-text ul li", func(i int, h *colly.HTMLElement) {
			if i > 0 {
				attribute := strings.Split(h.DOM.Text(), ":")
				attrName := strings.TrimSpace(attribute[0])
				attrValue := strings.TrimSuffix(strings.TrimSpace(strings.Title(strings.ToLower(attribute[1]))), ".")
				if len(attrValue) == 0 {
					return
				}
				if attrName == "Phong cách" || attrName == "Nhà pha chế" {
					values := strings.Split(attrValue, ",")
					for _, val := range values {
						normalizeVal := strings.TrimSpace(val)

						if strings.Contains(normalizeVal, "Và") {
							subAttr := strings.Split(normalizeVal, "Và")
							for _, subVal := range subAttr {
								attributes = append(attributes, fmt.Sprintf(`%s:%s`, attrName, translate(subVal)))
							}
						} else if strings.Contains(normalizeVal, "Nhưng") {
							subAttr := strings.Split(normalizeVal, "Nhưng")
							for _, subVal := range subAttr {
								attributes = append(attributes, fmt.Sprintf(`%s:%s`, attrName, translate(subVal)))
							}
						} else if strings.Contains(normalizeVal, "Đầy") {
							subAttr := strings.Split(normalizeVal, "Đầy")
							for _, subVal := range subAttr {
								attributes = append(attributes, fmt.Sprintf(`%s:%s`, attrName, translate(subVal)))
							}
						} else {

							attributes = append(attributes, fmt.Sprintf(`%s:%s`, attrName, translate(normalizeVal)))
						}
					}

				} else if attrName == "Nhóm hương" {
					values := strings.Split(attrValue, "-")
					for _, val := range values {
						attributes = append(attributes, fmt.Sprintf(`%s:%s`, attrName, strings.Title(strings.ToLower(strings.TrimSpace(val)))))
					}
				} else {
					attributes = append(attributes, fmt.Sprintf(`%s:%s`, attrName, attrValue))
				}

			}

		})
		product.facets = strings.Join(attributes, "|")
		summary, _ := e.DOM.Find(".product-detail-info .product .product-body-text .pb-20").Html()
		description, _ := e.DOM.Parent().Find(".block-perfume-info .row .info-left .desc").Html()
		product.description = fmt.Sprintf(`%s %s`, summary, description)
		variantCount := e.DOM.Find(".product-detail-info .product-detail-info-right .buy-product .buy-product-inner .product-top").Length()

		e.ForEach(".product-detail-info .product-detail-info-right .buy-product .buy-product-inner .product-top", func(i int, h *colly.HTMLElement) {

			productVariant := product
			productVariant.optionValues = strings.TrimSpace(h.DOM.Find(".product-name").Clone().Children().Remove().End().Text())
			productVariant.sku = h.DOM.Find(".product-name span .text-muted").Text()
			// if productVariant.optionValues == "" {
			// 	productVariant.optionGroups = ""
			// }
			if variantCount == 1 {
				productVariant.optionValues = ""
				productVariant.optionGroups = ""
			}

			// salePrice := h.DOM.Find(".price .sale .sale-price").Text()
			productVariant.price = specialRegex.ReplaceAllString(h.DOM.Find(".price").Clone().Children().Remove().End().Text(), "")
			if i > 0 {
				productVariant.assets = ""
				productVariant.facets = ""
				productVariant.optionGroups = ""
				productVariant.name = ""
				productVariant.slug = ""
				productVariant.description = ""
			}
			w.Write(productVariant.ToSlice())

		})
		// fmt.Println(summary)
		// fmt.Println(description)

	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})
	// "?page=1#price=0%252C0&gender=146&page=2&sort=new_arrival"

	for i := 0; i < info.limit; i++ {
		c.Visit(fmt.Sprintf("%s?page=%d", info.crawlUrl, i+1))
	}

}

type CrawlInfo struct {
	crawlUrl string
	filename string
	facets   string
	limit    int
}

func NewCrawl(typeCrawl string) {
	// Nu =====>
	// Nuoc hoa: https://www.thegioinuochoa.com.vn/nuoc-hoa-nu
	// Giftset: https://www.thegioinuochoa.com.vn/giftset-nu
	// Lan khu mui: https://www.thegioinuochoa.com.vn/lan-khu-mui-nu
	// My pham: https://www.thegioinuochoa.com.vn/my-pham-nu
	// Nam ====>
	// Nuoc hoa: https://www.thegioinuochoa.com.vn/nuoc-hoa-nam
	// Giftset: https://www.thegioinuochoa.com.vn/giftset-nam
	// Lan khu mui: https://www.thegioinuochoa.com.vn/lan-khu-mui-nam
	// My Pham: https://www.thegioinuochoa.com.vn/my-pham-nam
	// Unisex =====>
	// Nuoc hoa: https://www.thegioinuochoa.com.vn/nuoc-hoa-unisex
	// My pham: https://www.thegioinuochoa.com.vn/my-pham-unisex
	var crawlUrls []CrawlInfo = []CrawlInfo{
		{
			filename: "nuoc-hoa-nu.csv",
			facets:   "Sản phẩm:Nước hoa",
			limit:    11,
			crawlUrl: "https://www.thegioinuochoa.com.vn/nuoc-hoa-nu",
		},
		{
			filename: "giftset-nu.csv",
			facets:   "Sản phẩm:Giftset",
			limit:    2,
			crawlUrl: "https://www.thegioinuochoa.com.vn/giftset-nu#cate_id=136&price=0%2C0&gender=146&cate_id=136&page=5&sort=new_arrival",
		},
		{
			filename: "lan-khu-mui-nu.csv",
			facets:   "Sản phẩm:Lăn khử mùi",
			limit:    0,
			crawlUrl: "https://www.thegioinuochoa.com.vn/lan-khu-mui-nu",
		},
		{
			filename: "my-pham-nu.csv",
			facets:   "Sản phẩm:Mỹ phẩm",
			limit:    1,
			crawlUrl: "https://www.thegioinuochoa.com.vn/my-pham-nu",
		},
		// /// Nam
		{
			filename: "nuoc-hoa-nam.csv",
			facets:   "Sản phẩm:Nước hoa",
			limit:    6,
			crawlUrl: "https://www.thegioinuochoa.com.vn/nuoc-hoa-nam",
		},
		{
			filename: "giftset-nam.csv",
			facets:   "Sản phẩm:Giftset",
			limit:    2,
			crawlUrl: "https://www.thegioinuochoa.com.vn/giftset-nam",
		},
		{
			filename: "lan-khu-mui-nam.csv",
			facets:   "Sản phẩm:Lăn khử mùi",
			limit:    1,
			crawlUrl: "https://www.thegioinuochoa.com.vn/lan-khu-mui-nam",
		},
		{
			filename: "my-pham-nam.csv",
			facets:   "Sản phẩm:Mỹ phẩm",
			limit:    0,
			crawlUrl: "https://www.thegioinuochoa.com.vn/my-pham-nam",
		},
		// /// Unisex
		{
			filename: "nuoc-hoa-unisex.csv",
			facets:   "Sản phẩm:Nước hoa",
			limit:    2,
			crawlUrl: "https://www.thegioinuochoa.com.vn/nuoc-hoa-unisex",
		},
		{
			filename: "my-pham-unisex.csv",
			facets:   "Sản phẩm:Mỹ phẩm",
			limit:    1,
			crawlUrl: "https://www.thegioinuochoa.com.vn/my-pham-unisex",
		},
	}

	if typeCrawl == "all" {
		file, err := os.Create("products.csv")
		defer file.Close()
		if err != nil {

			log.Fatalln("failed to open file", err)
		}
		w := csv.NewWriter(file)
		headers := Product{}.GetHeader()
		w.Write(headers)
		for _, info := range crawlUrls {

			handleCrawl(info, w)
		}
	} else {
		for _, info := range crawlUrls {

			handleCrawl(info, nil)
		}
	}

}
