package main

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"log"
	"net/http"
	"time"
)

// createSchema creates database schema for User and Story models.
func createSchema(client *mongo.Client) error {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	defer client.Disconnect(ctx)
	client.Ping(ctx, readpref.Primary())
	log.Print("db is alive")

	networkdata := client.Database("network")
	channelsCollection := networkdata.Collection("channels")
	_, err := channelsCollection.InsertMany(ctx, []interface{}{
		bson.D{
			{
				Key:   "showrgb",
				Value: bson.A{80, 20, 60},
			},
			{Key: "description", Value: "Huge red, small green and a normal sized blue"},
		},
		bson.D{
			{
				Key:   "showrgb",
				Value: bson.A{20, 80, 60},
			},
			{Key: "description", Value: "Huge green, small red and a normal sized blue"},
		},
		bson.D{
			{
				Key:   "showrgb",
				Value: bson.A{20, 200, 60},
			},
			{Key: "description", Value: "ALL green, small red and a normal sized blue"},
		},
		bson.D{
			{
				Key:   "showrgb",
				Value: bson.A{20, 80, 300},
			},
			{Key: "description", Value: "ALL Blue, huge green, small red"},
		},
	})
	changeTo(client, 11) // start from channel 11

	return err
}

type Current struct {
	Channel int `bson:"channel"`
}

func changeTo(client *mongo.Client, n int) error {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	networkdata := client.Database("network")
	post := Current{
		Channel: n,
	}

	selectedChannelCollection := networkdata.Collection("selectedChannel")
	selectedChannelResult, err := selectedChannelCollection.InsertOne(ctx, post)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("selectedChannelResult: ", selectedChannelResult)

	return err
}

// this moves the channel and checking that it is always 0 <= n <= lastCh
func moveToChannel(lastCh int, n *int) {
	if *n > lastCh {
		*n = 0
	} else if *n < 0 {
		*n = lastCh - 1
	}
}

func lastChannel(client *mongo.Client) (error, int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	client.Ping(ctx, readpref.Primary())

	networkdata := client.Database("network")
	channelsCollection := networkdata.Collection("channels")
	totChs, err := channelsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	return err, int(totChs - 1)
}

// TODO: refactor this
// curl -v --request GET http://localhost:8080/channel/next
func next(w http.ResponseWriter, r *http.Request, n *int, totCh int, client *mongo.Client) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("serving Get /next ")
		*n++
		moveToChannel(totCh, n)
		changeTo(client, *n)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// TODO: refactor this
// curl -v --request GET http://localhost:8080/channel/previous
func previous(w http.ResponseWriter, r *http.Request, n *int, totCh int, client *mongo.Client) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("serving Get /previous ")
		*n--
		moveToChannel(totCh, n)
		changeTo(client, *n)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongostorage:27017/?authSource=admin&replicaSet=jamRS&directConnection=true"))
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) >= 2 {
		if os.Args[1] == "createDb" {
			log.Printf("creating database...")
			err := createSchema(client)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("unrecognised parameter: %v", os.Args[1])
		}
		return
	}

	log.Printf("starting channels api ")
	currentChannel := 1
	err, totCh := lastChannel(client)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("last channel is %v", totCh)

	http.HandleFunc("/channel/next",
		func(w http.ResponseWriter, r *http.Request) {
			next(w, r, &currentChannel, totCh, client)
		})
	http.HandleFunc("/channel/previous",
		func(w http.ResponseWriter, r *http.Request) {
			previous(w, r, &currentChannel, totCh, client)
		})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
