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
func image_stream(r float64, size_red float64, size_green float64, size_blue float64) []byte {
	var w, h int = 280, 240
	var hw, hh float64 = float64(w / 2), float64(h / 2)
	θ := 2 * math.Pi / 3
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

func iterateChangeStream(routineCtx context.Context, waitGroup sync.WaitGroup, stream *mongo.ChangeStream) {
	defer stream.Close(routineCtx)
	defer waitGroup.Done()
	for stream.Next(routineCtx) {
		var data bson.M
		if err := stream.Decode(&data); err != nil {
			panic(err)
		}
		fd := data["fullDocument"]
		fmt.Println(fd)
		fmt.Printf("type is %T", fd)

		username, ok := fd["username"].(string)

		for k, v := range data.(primitive.M)["fullDocument"] {
			if str, ok := v.(string); ok {
				fmt.Println(str)
				// Use k and str
			}
		}

		// fmt.Println(ok)
		// fmt.Println(md)

		// newChannel := data["channel"]
		// log.Printf("starting trasmission of channel %v\n", newChannel)

	}
}

func main() {
	log.Print("Database setting up ...")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/?authSource=admin&replicaSet=jamRS"))
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
	go iterateChangeStream(routineCtx, waitGroup, episodesStream)

	log.Printf("ANTENNA setting up...")
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Print("error upgrading ws ", err)
		}
		log.Print("ws upgraded (AKA new tv box connected)")

		// always on transmission on this channel
		go func() {
			defer conn.Close()
			n := 135.0
			incr := -0.5
			for {
				err = wsutil.WriteServerMessage(conn, ws.OpBinary, image_stream(n, 80, 20, 600)) // TODO: to be wired to the database
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
