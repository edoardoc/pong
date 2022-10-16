package main

/*

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
*/

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

// createSchema creates database schema for User and Story models.
func createSchema(client *mongo.Client) error {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := client.Connect(ctx)
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	log.Print("db is alive")

	networkdata := client.Database("network")
	channelsCollection := networkdata.Collection("channels")
	ccc, err := channelsCollection.InsertMany(ctx, []interface{}{
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
	fmt.Println("CHANNELS JUST SAVED: ", ccc)
	// cursorChannels := channelsCollection.FindOne(ctx, bson.M{"_id": 0})
	// if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer cursorChannels.Close(ctx)

	// 	fmt.Println("TOTAL: ", cursorChannels.count())

	// fmt.Println("CHANNEL 1:", cursorChannels)

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
	return err
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
	// http.HandleFunc("/api/signup", Signup)
	// http.HandleFunc("/api/login", Login)
	// http.HandleFunc("/api/users", Users)
	http.ListenAndServe(":8080", nil)
}