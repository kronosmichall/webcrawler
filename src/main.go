package crawler 

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func getWebsiteBody(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	// limitedReader := io.LimitReader(resp.Body, 2000000)
	data, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error", err)
		return nil, err 
	}	
	return (data), nil
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

type Result struct {
	url string
	depth uint
}

func extractUrlsAndPublish(path string, ch chan Result, depth uint) {
	body, err := getWebsiteBody(path)
	if err != nil {
		fmt.Println("Skipping site: ", path)
		return
	}

	urls := getUrls(body)
	for _,url := range urls {
		result := Result{string(url), depth+1}
		ch <- result
	}
}

func bfs(path string, depth uint, limit uint) {
	ch := make(chan Result)
	baseRefs := map[string]uint{}
	refs := map[string]uint{}
	
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	go func() {
		ch <- Result{path,0}
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
				go extractUrlsAndPublish(result.url, ch, result.depth)
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
	fmt.Println("Hello world")
	d := flag.Uint("d", 0, "depth of bfs search")
	l := flag.Uint("l", 10, "same base url limit of requests")
	s := flag.String("s", "", "starting url")
	flag.Parse()

	bfs(*s, *d, *l)
}
