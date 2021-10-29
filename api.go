package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/MemeLabs/dggchat"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (b *bot) sendtoDB(m dggchat.Message) error {
	_, err := b.mongo.Database("mentions").Collection(m.Sender.Nick).InsertOne(context.TODO(), m)
	if err != nil {
		return err
	}
	return nil
}

func (b *bot) requestMentions(m dggchat.Message, limit int64) ([]dggchat.Message, error) {
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(limit)
	collection := b.mongo.Database("mentions").Collection(m.Sender.Nick)

	cur, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Printf("[##] send error: %s\n", err.Error())
		return nil, err
	}

	var results []dggchat.Message
	for cur.Next(context.TODO()) {
		var record dggchat.Message
		err := cur.Decode(&record)
		if err != nil {
			log.Printf("[##] send error: %s\n", err.Error())
			return nil, err
		}

		results = append(results, record)
	}
	if err := cur.Err(); err != nil {
		log.Printf("[##] send error: %s\n", err.Error())
		return nil, err
	}

	cur.Close(context.TODO())

	return results, nil
}

func (b *bot) getOptedUsers() ([]string, error) {
	opts := options.Find().SetProjection(bson.M{"username": 1, "_id": 0})
	collection := b.mongo.Database("mentions").Collection("optedusers")

	cur, err := collection.Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		log.Printf("[##] send error: %s\n", err.Error())
		return nil, err
	}

	var results []bson.M
	err = cur.All(context.TODO(), &results)
	if err != nil {
		log.Printf("[##] send error: %s\n", err.Error())
		return nil, err
	}

	if err := cur.Err(); err != nil {
		log.Printf("[##] send error: %s\n", err.Error())
		return nil, err
	}

	cur.Close(context.TODO())

	var users []string
	for _, result := range results {
		users = append(users, result["username"].(string))
	}

	return users, nil
}

func UploadToHost(data []byte) (string, error) {
	postData := &bytes.Buffer{}
	mw := multipart.NewWriter(postData)

	w, err := mw.CreateFormFile("file", "page.html")
	if err != nil {
		return "error creating form file", err
	}
	if _, err := w.Write(data); err != nil {
		return "error writing to form file", err
	}

	if err := mw.Close(); err != nil {
		return "error closing multipart writer", err
	}

	req, err := http.NewRequest("POST", "https://0x0.st", postData)
	if err != nil {
		return "http.NewRequest failed", err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "http Client failed", err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		log.Println(bodyString)
		return strings.Replace(bodyString, "\n", "", -1), nil
	}

	log.Println(resp)

	return "", fmt.Errorf("bad status: %s", resp.Status)
}
