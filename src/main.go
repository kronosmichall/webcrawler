package main

import (
	"flag"
	"fmt"
	"net/url"
	"regexp"
)

type Result struct {
	parentUrl string
	url       string
	depth     uint
}

func createPage(path string, title string, body string) Page {
	textBody := getBodyText(body)
	return Page{
		bsonStr(path),
		bsonStr(title),
		bsonStr(textBody),
	}
}

func getUrls(body []byte) [][]byte {
	re := regexp.MustCompile(`https?://[^"'\\s<]+`)

	links := re.FindAll(body, -1)
	if links == nil {
		return [][]byte{}
	}
	return links
}

func getBaseUrl(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	return u.Host, err
}

func insertToNeo4j(url1 string, url2 string, insert func(string, string) error) {
	base1, err := getBaseUrl(url1)
	if err != nil {
		fmt.Println(err)
		return
	}
	base2, err := getBaseUrl(url2)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = insert(base1, base2)
	if err != nil {
		fmt.Println("insert", err)
	}
}

func extractUrlsAndPublish(
	path string,
	ch chan Result,
	depth uint,
	mongoInsert func(Page) error,
	neo4jInsert func(string, string) error,
) {
	body, err := getWebsiteBody(path)
	if err != nil {
		fmt.Println("Skipping site: ", path)
		return
	}
	// go mongoInsertToDb(path, body, mongoInsert)

	urls := getUrls(body)
	for _, url := range urls {
		 insertToNeo4j(path, string(url), neo4jInsert)
		result := Result{path, string(url), depth + 1}
		ch <- result
	}
}

func bfs(path string, depth uint, limit uint) {
	ch := make(chan Result)
	baseRefs := map[string]uint{}
	refs := map[string]uint{}
	neo4jInsert, neo4jCleanup, err := neo4jConnector()
	if err != nil {
		panic(err)
	}
	defer neo4jCleanup()
	mongoInsert, err := mongoConnector("web", "websites")
	if err != nil {
		panic(err)
	}

	go func() {
		ch <- Result{path, path, 0}
	}()

	for {
		select {
		case result := <-ch:
			baseUrl, err := getBaseUrl(result.url)
			if err != nil {
				continue
			}
			if refs[result.url] == 0 && baseRefs[baseUrl] < limit && result.depth < depth {
				refs[result.url] += 1
				fmt.Println("Visiting ", result.url)
				go extractUrlsAndPublish(result.url, ch, result.depth, mongoInsert, neo4jInsert)
			} else {
				refs[result.url] += 1
			}
			baseRefs[baseUrl] += 1
		default:
			continue
		}
	}
}

func main() {
	// time.Sleep(1000 * time.Second)
	fmt.Println("Hello world")
	d := flag.Uint("d", 1, "depth of bfs search")
	l := flag.Uint("l", 10, "same base url limit of requests")
	s := flag.String("s", "", "starting url")
	flag.Parse()

	bfs(*s, *d, *l)
}
