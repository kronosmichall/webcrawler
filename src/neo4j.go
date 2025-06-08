package main

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type (
	empty       struct{}
	set         map[string]empty
	Connections map[string]set
)

func (c Connections) addConnection(url1 string, url2 string) {
	if c.checkConnection(url1, url2) {
		return
	}

	s1, exists1 := c[url1]
	s2, exists2 := c[url2]
	if !exists1 {
		s1 = make(set)
		c[url1] = s1
	}

	if !exists2 {
		s2 = make(set)
		c[url2] = s2
	}
	s1[url2] = empty{}
	s2[url1] = empty{}
}

func (c Connections) checkConnection(url1 string, url2 string) bool {
	s1, exists1 := c[url1]
	if !exists1 {
		return false
	}

	_, exists2 := s1[url2]
	return exists2
}

func neo4jConnector() (func(string, string) error, func(), error) {
	uri := "bolt://neo4j:7687"
	username := "neo4j"
	password := "12345678"

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	insert := func(url1 string, url2 string) error {
		query := ` 
			MERGE (w1:Website {url: $url1})
			MERGE (w2:Website {url: $url2})
			MERGE (w1)-[:CONNECTS_TO]-(w2)
			`
		params := map[string]any{
			"url1": url1,
			"url2": url2,
		}
		insertCtx, insertCancel := context.WithTimeout(ctx, 5 * time.Second)
		defer insertCancel()

		result, err := session.Run(insertCtx, query, params)
		if err != nil {
			return err
		}
		_, err = result.Consume(ctx)
		return err
	}

	cleanup := func() {
		driver.Close(ctx)
		session.Close(ctx)
	}
	return insert, cleanup, nil
}
