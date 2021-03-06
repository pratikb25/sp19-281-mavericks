package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	_ "github.com/streadway/amqp"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"os"
	"time"
	"bytes"
)

/*TODO: Support environment variables for Mongo Config*/

var mongodbServer = os.Getenv("MONGO_SERVER") + ":27017"
var mongodbDatabase = os.Getenv("DATABASE")
var mongodbCollection = os.Getenv("COLLECTION")
var dashboard_url = os.Getenv("DASHBOARD_URL")

func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})
	n := negroni.Classic()
	mx := mux.NewRouter()
	initRoutes(mx, formatter)
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	n.UseHandler(handlers.CORS(allowedHeaders, allowedMethods, allowedOrigins)(mx))
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/events", postEventHandler(formatter)).Methods("POST")
	mx.HandleFunc("/events", getAllEventsHandler(formatter)).Methods("GET")
	mx.HandleFunc("/events/{eventId}", getEventHandler(formatter)).Methods("GET")
	mx.HandleFunc("/events/{eventId}", deleteEventhandler(formatter)).Methods("DELETE")
	mx.HandleFunc("/events", optionsHandler(formatter)).Methods("OPTIONS")
	mx.HandleFunc("/events/{eventId}", optionsHandler(formatter)).Methods("OPTIONS")
}

/* Send new event data to dashboard */
func makeRequest(e *ScheduledEvent) {
	url := dashboard_url

	// Construct the message to be sent in request body
	message := map[string]interface{}{
		"bucket":    "eventbrite",
		"user_uuid": e.OrgId,
		"eventId":   e.EventId,
		"eventName": e.EventName,
		"location":  e.Location,
		"date":      e.Date,
	}
	

	// Marshal message into JSON format
	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
		return
	}

	// POST the message to dashboard_url
	_, err = http.Post(url, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		fmt.Println("Sent new event details to dashboard")
		log.Fatal(err)
		return
	}
}

/*TODO: Connect to MongoDb only when admin user is provided*/

func postEventHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var e EventPayload
		// Open MongoDB Session
		_ = json.NewDecoder(req.Body).Decode(&e)
		fmt.Println("Event: ", e)
		session, err := mgo.Dial(mongodbServer)
		if err != nil {
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodbDatabase).C(mongodbCollection)
		var match EventPayload
		// TODO: Revisit correct format for sending date
		//localDate, err := time.Parse("2006-01-02", e.Date)
		localDate, err := time.Parse("2006-01-02", e.Date)
		err = c.Find(bson.M{"eventName": e.EventName}).One(&match)
		fmt.Println("Match: ", match)
		eventId, _ := uuid.NewV4()

		if err == nil {
			fmt.Printf("Event %s is already scheduled with the same name provided!", match.EventName)
		} else {
			eventEntry := ScheduledEvent{
				EventId:    eventId.String(),
				EventName:  e.EventName,
				Date:       localDate,
				Location:   e.Location,
				OrgId:      e.OrgId,
				BucketName: e.BucketName,
				Price:      e.Price}

			err = c.Insert(eventEntry)

			if err != nil {
				fmt.Println("Error while adding Events: ", err)
				formatter.JSON(w, http.StatusInternalServerError,
					struct{ Response error }{err})
			} else {
				formatter.JSON(w, http.StatusOK,
					struct{ Response string }{"Event successfully added"})
				/* POST new event details to the dashboard */
				makeRequest(&eventEntry)
			}

		}
	}
}

func optionsHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		setupResponse(&w, req)
		fmt.Println("options handler PREFLIGHT Request")
		return
	}
}


func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}


// API Get All Events Handler
func getEventHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		setupResponse(&w, req)
		session, err := mgo.Dial(mongodbServer)
		if err != nil {
			panic(err)
		}

		if err != nil {
			fmt.Println("Events API (Get) - Unable to connect to MongoDB during read operation")
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodbDatabase).C(mongodbCollection)

		var results []ScheduledEvent
		params := mux.Vars(req)
		var eventId string = params["eventId"]
		fmt.Printf("Event ID: %s", eventId)
		err = c.Find(bson.M{"eventId": eventId}).All(&results)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(results)
		response := EventResponse{
			Count:  len(results),
			Events: results}

		if len(results) > 0 {
			//formatter.JSON(w, http.StatusOK, response)
			formatter.JSON(w, http.StatusOK, response)
		} else {
			formatter.JSON(w, http.StatusNoContent,
				struct{ Response string }{"No Events found for the given ID"})
		}
	}
}

// API Get All Events Handler
func getAllEventsHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		setupResponse(&w, req)
		fmt.Printf("Server is : %s", mongodbServer)
		session, err := mgo.Dial(mongodbServer)
		if err != nil {
			panic(err)
		}

		if err != nil {
			fmt.Println("Events API (Get) - Unable to connect to MongoDB during read operation")
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodbDatabase).C(mongodbCollection)
		var results []ScheduledEvent
		err = c.Find(nil).All(&results)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(results)
		response := EventResponse{
			Count:  len(results),
			Events: results}
		if len(results) > 0 {
			formatter.JSON(w, http.StatusOK, response)
		} else {
			formatter.JSON(w, http.StatusNoContent,
				struct{ Response string }{"No Events found"})
		}
	}
}

func deleteEventhandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		session, err := mgo.Dial(mongodbServer)
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodbDatabase).C(mongodbCollection)
		params := mux.Vars(req)
		var eventId string = params["eventId"]
		fmt.Println("Event To Delete is: ", eventId)
		err = c.Remove(bson.M{"eventId": eventId})
		if err != nil {
			fmt.Printf("Event not found with ID: %s", eventId)
			formatter.JSON(w, http.StatusNotFound, "Event Not Found")
			return
		}
		formatter.JSON(w, http.StatusOK, "Event: "+
			eventId+" Deleted")
	}
}

// Helper Functions
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// API Ping Handler
func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("Server is : %s", mongodbServer)
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"API version 1.0 alive!"})
	}
}

