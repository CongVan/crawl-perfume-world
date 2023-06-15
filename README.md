## Crawl perfume world website data
- Target page: https://www.thegioinuochoa.com.vn/
- Output: all product data follow [Vendure format](https://github.com/vendure-ecommerce/vendure/blob/master/packages/core/mock-data/data-sources/products.csv)

## Scripts
```sh
  go run main.go crawl # Save each category to each file
  go run main.go crawl all # Save each category to one file
```