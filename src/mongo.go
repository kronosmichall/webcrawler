package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	MaxBsonSize  = 1_000_000
	InsertTimout = 5 * time.Second
)

func bsonStr(str string) string {
	if len(str) <= MaxBsonSize {
		return str
	} else {
		return str[:MaxBsonSize]
	}
}

type Page struct {
	URL   string `bson:"url"`
	Title string `bson:"title"`
	Body  string `bson:"body"`
}

func mongoInsertToDb(path string, data []byte, insert func(Page) error) error {
	// if !utf8.Valid(data) {
	// 	return errors.New("data is not valid utf-8")
	// }
	title := getTitleOrH1(data)
	page := createPage(path, title, string(data))
	return insert(page)
}

func mongoConnector(db string, collection string) (func(Page) error, error) {
	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		fmt.Println("Could not find uri in env; fallback to localhost")
		uri = "mongodb://127.0.0.1:27017"
	} else {
		fmt.Printf("getting uri from env: %s\n", uri)
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	c := client.Database(db).Collection(collection)
	insert := func(page Page) error {
		ctx, cancel := context.WithTimeout(context.Background(), InsertTimout)
		defer cancel()
		_, err := c.InsertOne(ctx, page)
		return err
	}
	return insert, nil
}
