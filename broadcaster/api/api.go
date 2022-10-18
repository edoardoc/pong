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
	if *n >= lastCh {
		*n = 0
	} else if *n <= 0 {
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

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/?authSource=admin&replicaSet=jamRS&directConnection=true"))
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
	var i int
	i = 1
	err, totCh := lastChannel(client)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("last channel is %v", totCh)

	i = 19
	moveToChannel(totCh, &i)
	i++
	moveToChannel(totCh, &i)
	i++
	moveToChannel(totCh, &i)
	i++
	moveToChannel(totCh, &i)
	i++
	moveToChannel(totCh, &i)
	i++
	moveToChannel(totCh, &i)
	log.Print("we are in channel ", i)

	changeTo(client, 78)
	// SAMPLE CODE TO SHOW ALL CHANNELS
	// channelsCollection := networkdata.Collection("channels")
	// cursor, err := channelsCollection.Find(ctx, bson.M{})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// var channels []bson.M
	// if err = cursor.All(ctx, &channels); err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(channels)

	// SAMPLE CODE TO SHOW CHANNEL with ID:...
	// objectId, err := primitive.ObjectIDFromHex("634bfe67eb5543ddd0dcc82b")
	// if err != nil {
	// 	log.Println("Invalid id")
	// }
	// channelsCollection := networkdata.Collection("channels")
	// cursorChannels := channelsCollection.FindOne(ctx, bson.M{"_id": objectId})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// var channel bson.M
	// if err = cursorChannels.Decode(&channel); err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(channel)

	// SAMPLE CODE TO SHOW CHANNEL Nth
	// channelsCollection := networkdata.Collection("channels")
	// options := new(options.FindOptions)
	// options.SetSkip(15)
	// cursor, err := channelsCollection.Find(ctx, bson.M{}, options)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer cursor.Close(ctx)
	// cursor.TryNext(ctx)
	// var result bson.M
	// if err := cursor.Decode(&result); err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(result["showrgb"])
	// transmission := result["showrgb"]
	// if pa, ok := transmission.(primitive.A); ok {
	// 	transmissionMSI := []interface{}(pa)
	// 	fmt.Println("Working", transmissionMSI)
	// 	fmt.Println(reflect.TypeOf(transmissionMSI))
	// }

	// transmission := []interface{}(result["showrgb"])
	// fmt.Printf("%T", transmission)

	// fmt.Println("CHANNEL 1:", cursorChannels)

	// http.HandleFunc("/api/signup", Signup)
	// http.HandleFunc("/api/login", Login)
	// http.HandleFunc("/api/users", Users)
	http.ListenAndServe(":8080", nil)
}
