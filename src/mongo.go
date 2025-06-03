package main

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/v2/mongo"
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

func insertToDb(path string, data []byte, client *mongo.Client) error {
	if !utf8.Valid(data) {
		return errors.New("data is not valid utf-8")
	}
	title := getTitleOrH1(data)
	page := createPage(path, title, string(data))
	collection := client.Database("web").Collection("websites")
	ctx, cancel := context.WithTimeout(context.Background(), InsertTimout)
	defer cancel()

	_, err := collection.InsertOne(ctx, page)
	return err
}
