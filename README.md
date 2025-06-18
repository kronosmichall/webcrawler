# Go webcrawler 

This Go-based webcrawler uses a **Breadth-First Search (BFS)** strategy to crawl the web.

### Features
 - Parses web pages and discovers linked URLs using BFS
 - Limits requests to any single domain to avoid overloading servers
 - Stores crawled data in MongoDB
 - Records URL connections in Neo4j for graph-based analysis

### Usage
The crawler requires three command-line arguments:
 - `-d` : Maximum depth for BFS traversal
 - `-l` : Limit of requests per domain (helps avoid spamming frequently referenced domains)
 - `-s` : Starting URL

### Data storage
**MongoDB**
Each crawled page is stored in MongoDB with the following fields:
```go
URL   string
Title string
Body  string
```

**Neo4j**
Discovered links between web pages are saved in Neo4j, enabling powerful graph-based analysis.
```go
(w1:Website {url: $url1})-[:CONNECTS_TO]-(w2:Website {url: $url2})
```
