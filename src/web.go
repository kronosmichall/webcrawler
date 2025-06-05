package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

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

func getText(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}

	text := []string{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text = append(text, getText(c))
	}
	return strings.Join(text, "\n")
}

func getBodyNode(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if body := getBodyNode(c); body != nil {
			return body
		}
	}
	return nil
}

func getBodyText(body string) string {
	doc, err := html.Parse(bytes.NewBufferString(body))
	if err != nil {
		return ""
	}
	bodyNode := getBodyNode(doc)
	text := getText(bodyNode)

	return NormalizeWhitespace(text)
}

var spaceRegex = regexp.MustCompile(`\s+`)

func NormalizeWhitespace(s string) string {
	trimmed := strings.TrimSpace(s)
	return spaceRegex.ReplaceAllString(trimmed, " ")
}
