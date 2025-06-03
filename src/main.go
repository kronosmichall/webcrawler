package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/net/html"
)

type Result struct {
	url   string
	depth uint
}
type Page struct {
	URL   string `bson:"url"`
	Title string `bson:"title"`
	Body  string `bson:"body"`
}

func getWebsiteBody(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}

	return data, nil
}

func insertToDb(path string, data []byte, client *mongo.Client) error {
	title := getTitleOrH1(data)
	page := Page{path, title, string(data)}
	collection := client.Database("web").Collection("websites")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, page)
	return err
}

func getTitleOrH1(body []byte) string {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return ""
	}

	var title string
	var h1 string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "title" && title == "" && n.FirstChild != nil {
				title = strings.TrimSpace(n.FirstChild.Data)
			}
			if n.Data == "h1" && h1 == "" && n.FirstChild != nil {
				h1 = strings.TrimSpace(n.FirstChild.Data)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if title != "" {
		return title
	}
	return h1
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

func extractUrlsAndPublish(path string, ch chan Result, depth uint, client *mongo.Client) {
	body, err := getWebsiteBody(path)
	if err != nil {
		fmt.Println("Skipping site: ", path)
		return
	}
	err = insertToDb(path, body, client)
	if err != nil {
		panic(err)
	}

	urls := getUrls(body)
	for _, url := range urls {
		result := Result{string(url), depth + 1}
		ch <- result
	}
}

func bfs(path string, depth uint, limit uint) {
	ch := make(chan Result)
	baseRefs := map[string]uint{}
	refs := map[string]uint{}

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		fmt.Println("Could not find uri in env; fallback to localhost")
		uri = "mongodb://127.0.0.1:27017"
	} else {
		fmt.Printf("getting uri from env: %s\n", uri)
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	go func() {
		ch <- Result{path, 0}
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
				go extractUrlsAndPublish(result.url, ch, result.depth, client)
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
