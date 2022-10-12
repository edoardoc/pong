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
					fmt.Println(err)
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
			fmt.Println("creating database...")
			// ** DB CREATION
			err := createSchema(db)
			if err != nil {
				log.Fatal(err)
			}
		} else if os.Args[1] == "dropDb" {
			fmt.Println("N/D")
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
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
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

func image_stream() []byte {
	var w, h int = 280, 240
	var hw, hh float64 = float64(w / 2), float64(h / 2)
	r := 40.0
	θ := 2 * math.Pi / 3
	cr := &Circle{hw - r*math.Sin(0), hh - r*math.Cos(0), 60}
	cg := &Circle{hw - r*math.Sin(θ), hh - r*math.Cos(θ), 60}
	cb := &Circle{hw - r*math.Sin(-θ), hh - r*math.Cos(-θ), 60}

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

	// var byteWriter = bufio.NewWriter(send_S3)

	buf := new(bytes.Buffer)
	png.Encode(buf, m)
	// f, err := os.OpenFile("rgb.png", os.O_WRONLY|os.O_CREATE, 0600)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer f.Close()
	// png.Encode(f, m)
	return buf.Bytes()
}

func main() {
	fmt.Println("starting echo...")

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("setting up...")
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			fmt.Println("upgrading ws ", err)

		}
		fmt.Println("ws upgraded")
		go func() {
			defer conn.Close()

			var (
				state  = ws.StateServerSide
				reader = wsutil.NewReader(conn, state)
				writer = wsutil.NewWriter(conn, state, ws.OpBinary)
			)
			for {
				header, err := reader.NextFrame()
				if err != nil {
					// handle error
				}

				// Reset writer to write frame with right operation code.
				writer.Reset(conn, state, header.OpCode)

				wsutil.WriteServerMessage(conn, ws.OpBinary, image_stream())

				// if _, err = io.Copy(writer, reader); err != nil {
				// 	// handle error
				// }
				if err = writer.Flush(); err != nil {
					// handle error
				}
			}
		}()
	}))

	// http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	conn, _, _, _ := ws.UpgradeHTTP(r, w)

	// 	go func() {
	// 		defer conn.Close()

	// 		var msg []byte
	// 		var op ws.OpCode
	// 		for {
	// 			msg, op, _ = wsutil.ReadClientData(conn)
	// 			if msg != nil {
	// 				break
	// 			}
	// 		}

	// 		fmt.Println("someone connected, starting transmission...") // needs at least one client to connect to start transmissions
	// 		for {
	// 			// msg := fmt.Sprintf("currentTimeMillis = %d", time.Now().UnixNano()/int64(time.Millisecond)) // simple stub transmission of local server timestamp
	// 			// fmt.Println("I AM sending ", msg)
	// 			// wsutil.WriteServerMessage(conn, op, []byte(msg))

	// 			wsutil.WriteServerMessage(conn, op, image_stream())
	// 		}
	// 	}()
	// }))
}
