package main

import (
	"craw-perfume-world/crawl"
	"craw-perfume-world/sanitize"
	"os"
)

func main() {
	action := os.Args[1]

	var typeCrawl = ""
	if len(os.Args) > 2 {
		typeCrawl = os.Args[2]
	}

	if action == "crawl" {
		crawl.NewCrawl(typeCrawl)
	}

	if action == "sanitize" {
		// option2 code executes here.
		sanitize.UpdateDescription()
	}

}
