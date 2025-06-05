package main

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
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
	uri := "neo4j://neo4j:7687"
	username := "neo4j"
	password := "12345678"

	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, nil, err
	}

	session, err := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	if err != nil {
		return nil, nil, err
	}

	// connections := Connections{}

	insert := func(url1 string, url2 string) error {
		// if connections.checkConnection(url1, url2) {
		// 	return nil
		// } else {
		// 	connections.addConnection(url1, url2)
		// }

		query := ` 
			MERGE (w1:Website) {url: $url1}
			MERGE (w2:Website) {url: $url2}
			MERGE (w1)-[:CONNECTS_TO]-(w2)
			`
		params := map[string]any{
			"url1": url1,
			"url2": url2,
		}
		_, err := session.Run(query, params)
		return err
	}

	cleanup := func() {
		driver.Close()
		session.Close()
	}
	return insert, cleanup, nil
}
