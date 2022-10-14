package main

/*

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var (
	signKey = []byte("JKJKKJKLtestingwaat")
	db      *bun.DB
)

type Apiuser struct {
	Email     string `bun:",pk"`
	Password  string `json:"password,omitempty"`
	Firstname string
	Lastname  string
}
type ApiUserClaims struct {
	*jwt.StandardClaims
	Apiuser
}

// createSchema creates database schema for User and Story models.
func createSchema(db *bun.DB) error {
	ctx := context.Background()

	models := []interface{}{
		(*Apiuser)(nil),
	}

	for _, model := range models {
		log.Printf("creating table %v ", reflect.TypeOf(model))

		_, err := db.NewCreateTable().Model(model).Exec(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func createToken(email string) (time.Time, string, error) {
	expirationTime := time.Now().Add(30 * time.Minute)
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &ApiUserClaims{
		&jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
		Apiuser{Email: email},
	}
	newtoken, err := t.SignedString(signKey)
	return expirationTime, newtoken, err
}

func decodeToken(tokenString string) (string, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &ApiUserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return signKey, nil
	})
	claims := token.Claims.(*ApiUserClaims)
	return claims.Apiuser.Email, err
}

func saveApiUser(item Apiuser) error {
	ctx := context.Background()
	_, err := db.NewInsert().
		Model(&item).
		On("CONFLICT (email) DO UPDATE").
		Set("Firstname = EXCLUDED.Firstname").
		Set("Lastname = EXCLUDED.Lastname").
		Exec(ctx)
	return err
}

func updateApiUser(item Apiuser) error {
	ctx := context.Background()
	_, err := db.NewUpdate().
		Model(&item).
		Set("firstname = ?", item.Firstname).
		Set("lastname = ?", item.Lastname).
		WherePK().
		// Where("email = ?", item.Email).
		Exec(ctx)
	return err
}

func checkLoginApiUser(item Apiuser) (bool, error) {
	ctx := context.Background()
	count, err := db.NewSelect().
		Model(&item).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("email = ?", item.Email).Where("password = ?", item.Password)
		}).
		Count(ctx)
	return count == 1, err
}

func receivedUser(r *http.Request) (Apiuser, error) {
	decoder := json.NewDecoder(r.Body)
	var oneUser Apiuser
	err := decoder.Decode(&oneUser)
	return oneUser, err
}

// curl -v --header "Content-Type: application/json" --request POST --data '{"email":"ridleys@gmail.com","password":"233223edfsdf","firstname":"Edoardo","lastname":"Ceccarelli"}' http://localhost:8080/api/signup
func Signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		log.Printf("serving /Signup ")

		oneUser, err := receivedUser(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			// a valid user was received (fair assumption, more checks are needed)

			// ** TOKEN GENERATION
			_, tokenString, err := createToken(oneUser.Email)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				err = saveApiUser(oneUser)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("x-auth-token", tokenString)
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// curl -v --header "Content-Type: application/json" --request POST --data '{"email":"ridleys@gmail.com","password":"test123"}' http://localhost:8080/api/login
func Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		log.Printf("serving /Login ")

		oneUser, err := receivedUser(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			credentialsOk, err := checkLoginApiUser(oneUser)
			if credentialsOk && err == nil {
				// ** TOKEN GENERATION
				_, tokenString, err := createToken(oneUser.Email)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("x-auth-token", tokenString)
					w.WriteHeader(http.StatusOK)
				}
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// curl -v --header "X-Auth-Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDYwNjU2NDgsIkVtYWlsIjoicmlkbGV5c0BnbWFpbC5jb20iLCJQYXNzd29yZCI6IiIsIkZpcnN0bmFtZSI6IiIsIkxhc3RuYW1lIjoiIn0.wbJl8b1xjsTavzk8g4mumDOt3NROHXv8Z-AoCBG1tvM" --header "Content-Type: application/json" --request PUT --data '{"firstname":"wwwEdoardo","lastname":"Ceccadddrelli"}' http://localhost:8080/api/users
// curl -v --header "X-Auth-Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDYwNjU2NDgsIkVtYWlsIjoicmlkbGV5c0BnbWFpbC5jb20iLCJQYXNzd29yZCI6IiIsIkZpcnN0bmFtZSI6IiIsIkxhc3RuYW1lIjoiIn0.wbJl8b1xjsTavzk8g4mumDOt3NROHXv8Z-AoCBG1tvM" --header "Content-Type: application/json" --request GET http://localhost:8080/api/users
func Users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("serving Get /Users ")
		receivedToken := r.Header.Get("x-auth-token")
		_, err := decodeToken(receivedToken) // ** a valid token means ok to give the list
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			// ** Apiuser LIST USERS
			ctx := context.Background()
			var listUser []*Apiuser
			err = db.NewSelect().
				Model((*Apiuser)(nil)).
				ColumnExpr("email").
				ColumnExpr("firstname").
				ColumnExpr("lastname").
				Scan(ctx, &listUser)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				formattedOut, err := json.MarshalIndent(listUser, "", "    ")
				// formattati, err := json.Marshal(listUser)
				if err != nil {
					log.Printf(err)
					return
				}
				w.Write(formattedOut)
				w.WriteHeader(http.StatusOK)
			}
		}
	case http.MethodPut:
		log.Printf("serving Put /Users ")
		receivedToken := r.Header.Get("x-auth-token")
		log.Printf("receivedToken : %v\n", receivedToken)
		whichUser, err := decodeToken(receivedToken) // ** DECODING TOKEN
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			oneUser, err := receivedUser(r)
			oneUser.Email = whichUser // to make sure the seek happens on the token value
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				err = updateApiUser(oneUser)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	// ** DB CONNECTION
	dsn := "postgres://waatusr:@database:5432/waat?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db = bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	if len(os.Args) >= 2 {
		if os.Args[1] == "createDb" {
			log.Printf("creating database...")
			// ** DB CREATION
			err := createSchema(db)
			if err != nil {
				log.Fatal(err)
			}
		} else if os.Args[1] == "dropDb" {
			log.Printf("N/D")
		} else {
			fmt.Printf("unrecognised parameter: %v", os.Args[1])
		}
		return
	}

	log.Printf("starting api ")
	http.HandleFunc("/api/signup", Signup)
	http.HandleFunc("/api/login", Login)
	http.HandleFunc("/api/users", Users)
	http.ListenAndServe(":8080", nil)
}


*/

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	// "go.mongodb.org/mongo-driver/mongo/readpref"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"net/http"
	"time"
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

// // curl -v --header "Content-Type: application/json" --request PUT --data '{"firstname":"wwwEdoardo","lastname":"Ceccadddrelli"}' http://localhost:8080/api/users
// func Users(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case http.MethodPut:
// 		log.Printf("serving Put /Users ")
// 		receivedToken := r.Header.Get("x-auth-token")
// 		log.Printf("receivedToken : %v\n", receivedToken)
// 		whichUser, err := decodeToken(receivedToken) // ** DECODING TOKEN
// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 		} else {
// 			oneUser, err := receivedUser(r)
// 			oneUser.Email = whichUser // to make sure the seek happens on the token value
// 			if err != nil {
// 				w.WriteHeader(http.StatusBadRequest)
// 			} else {
// 				err = updateApiUser(oneUser)
// 				if err != nil {
// 					w.WriteHeader(http.StatusInternalServerError)
// 				} else {
// 					w.WriteHeader(http.StatusOK)
// 				}
// 			}
// 		}
// 	default:
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	}
// }

// info about scaling... https://medium.com/free-code-camp/million-websockets-and-go-cc58418460bb
func main() {
	log.Printf("setting up...")

	// DB CODE STARTS HERE
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://jamgg:jam@localhost:27017/?authSource=admin"))
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
	log.Print("connected to db")

	networkdata := client.Database("network")

	channelsCollection := networkdata.Collection("channels")
	channelsResult, err := channelsCollection.InsertMany(ctx, []interface{}{
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
	if err != nil {
		log.Fatal(err)
	}

	cursorChannels := channelsCollection.FindOne(ctx, bson.M{"_id": 0})
	if err != nil {
		log.Fatal(err)
	}
	// defer cursorChannels.Close(ctx)

	fmt.Println("CHANNEL 1:", cursorChannels)

	// selectedChannelCollection := networkdata.Collection("selectedChannel")
	// selectedChannelResult, err := selectedChannelCollection.InsertOne(ctx, bson.D{
	// 	{Key: "channel", Value: cursorChannels.ID()},
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("selectedChannelResult: ", selectedChannelResult)

	// Now I open

	// for cursor.Next(ctx) {
	// 	fmt.Println("cursor:", cursor.Current)
	// 	var channel bson.M
	// 	if err = cursor.Decode(&channel); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	//		fmt.Println("CANALE:", channel["_id"])
	// }

	fmt.Printf("%v ready for transmission!\n", len(channelsResult.InsertedIDs))
	// DB CODE ENDS HERE

	// http.HandleFunc("/api/users", Users)
	// http.ListenAndServe(":8080", nil)

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Printf("error upgrading ws ", err)
		}
		log.Printf("ws upgraded (AKA new tv box connected)")

		// always on transmission on this channel
		go func() {
			defer conn.Close()
			n := 135.0
			incr := -0.5
			for {
				err = wsutil.WriteServerMessage(conn, ws.OpBinary, image_stream(n, 80, 20, 600))
				if err != nil {
					log.Printf("tv shut down, stop this feed ", err)
					break
				}
				n = n + incr
				if n > 135 || n < 20 {
					incr = -incr
				}
			}
			log.Printf("ADIOS")
		}()
	}))
}
