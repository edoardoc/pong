package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Circle struct {
	X, Y, R float64
}

func (c *Circle) Brightness(x, y float64) uint8 {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		// outside
		return 0
	} else {
		// inside
		return uint8((1 - math.Pow(d, 5)) * 255)
	}
}

// from http://tech.nitoyon.com/en/blog/2015/12/31/go-image-gen/
func image_stream(r float64, transmission [3]int) []byte {
	var w, h int = 280, 240
	var hw, hh float64 = float64(w / 2), float64(h / 2)
	θ := 2 * math.Pi / 3

	size_red := float64(transmission[0])
	size_green := float64(transmission[1])
	size_blue := float64(transmission[2])

	cr := &Circle{hw - r*math.Sin(0), hh - r*math.Cos(0), size_red}
	cg := &Circle{hw - r*math.Sin(θ), hh - r*math.Cos(θ), size_green}
	cb := &Circle{hw - r*math.Sin(-θ), hh - r*math.Cos(-θ), size_blue}

	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c := color.RGBA{
				cr.Brightness(float64(x), float64(y)),
				cg.Brightness(float64(x), float64(y)),
				cb.Brightness(float64(x), float64(y)),
				255,
			}
			m.Set(x, y, c)
		}
	}

	buf := new(bytes.Buffer)
	png.Encode(buf, m)
	return buf.Bytes()
}

type DbEvent struct {
	DocumentKey   documentKey  `bson:"documentKey"`
	OperationType string       `bson:"operationType"`
	FullDocument  fullDocument `bson:"fullDocument"`
}
type documentKey struct {
	ID primitive.ObjectID `bson:"_id"`
}

type fullDocument struct {
	ID      primitive.ObjectID `bson:"_id"`
	Channel int                `bson:"channel"`
}

// this one gets the the nth channel transmission data (3 values)
// TODO: transmissionOfChannel is fragile, upon receiving wrong n it crashes!!!!
func transmissionOfChannel(client *mongo.Client, n int) []interface{} {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Connect(ctx)
	networkdata := client.Database("network")
	channelsCollection := networkdata.Collection("channels")
	options := new(options.FindOptions)

	options.SetSkip(int64(n))
	cursor, err := channelsCollection.Find(ctx, bson.M{}, options)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)
	cursor.TryNext(ctx)
	var result bson.M
	if err := cursor.Decode(&result); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(result["showrgb"])
	transmission := result["showrgb"]
	if pa, ok := transmission.(primitive.A); ok {
		transmissionMSI := []interface{}(pa)
		fmt.Println("Working", transmissionMSI)
		return transmissionMSI
	}
	return nil
}

func iterateChangeStream(client *mongo.Client, routineCtx context.Context, waitGroup sync.WaitGroup, stream *mongo.ChangeStream, trn *[3]int) {
	defer stream.Close(routineCtx)
	defer waitGroup.Done()
	for stream.Next(routineCtx) {

		var dbe DbEvent
		if err := stream.Decode(&dbe); err != nil {
			panic(err)
		}
		fmt.Printf("Channel: %v\n", dbe.FullDocument.Channel)
		newones := transmissionOfChannel(client, dbe.FullDocument.Channel)
		(*trn)[0] = int(newones[0].(int32))
		(*trn)[1] = int(newones[1].(int32))
		(*trn)[2] = int(newones[2].(int32))
		fmt.Printf("TRANSMITTING ACTUAL: %v\n", *trn)
	}
}

func main() {
	transmission := [3]int{78, 89, 45} // this is the data (generator) for a channel transmission, starting up with any transmission

	log.Print("Database setting up ...")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongostorage:27017/?authSource=admin&replicaSet=jamRS&directConnection=true"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	log.Print("connected to db and pinged")

	// watching for channel changes...
	networkdata := client.Database("network")
	selectedChannelCollection := networkdata.Collection("selectedChannel")
	var waitGroup sync.WaitGroup
	episodesStream, err := selectedChannelCollection.Watch(context.TODO(), mongo.Pipeline{})
	if err != nil {
		panic(err)
	}
	waitGroup.Add(1)
	routineCtx, _ := context.WithCancel(context.Background())
	go iterateChangeStream(client, routineCtx, waitGroup, episodesStream, &transmission)

	log.Printf("ANTENNA setting up...")
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Print("error upgrading ws ", err)
		}
		log.Print("ws upgraded (AKA new tv box connected)")

		// always on transmission(s) on this channel
		go func() {
			defer conn.Close()
			n := 135.0
			incr := -0.5
			for {
				err = wsutil.WriteServerMessage(conn, ws.OpBinary, image_stream(n, transmission)) // wired to the database
				if err != nil {
					log.Print("tv shut down, stop this feed ", err)
					break
				}
				n = n + incr
				if n > 135 || n < 20 {
					incr = -incr
				}
			}
			log.Print("ADIOS")
		}()
	}))
	waitGroup.Wait()
}
