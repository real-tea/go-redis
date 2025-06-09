package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// App holds the application's dependencies, like the router and database client.
type App struct {
	Router *mux.Router
	RDB    *redis.Client
}

// A simple struct for our JSON response
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// initialize sets up the database connection and the router.
func (a *App) initialize() {
	// Get Redis host from environment variable, default to localhost if not set.
	// This is key for switching between local Docker and AWS ElastiCache.
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost:6379"
	}

	// Initialize Redis client
	a.RDB = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Check the connection
	ctx := context.Background()
	_, err := a.RDB.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	fmt.Println("Successfully connected to Redis.")

	// Initialize the router
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

// initializeRoutes defines the API endpoints.
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/get/{key}", a.getValue).Methods("GET")
	a.Router.HandleFunc("/set", a.setValue).Methods("POST")
}

// run starts the HTTP server.
func (a *App) run(addr string) {
	log.Printf("Server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// respondWithError sends a JSON error message.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// getValue handles GET requests to retrieve a value from Redis.
func (a *App) getValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	ctx := context.Background()

	val, err := a.RDB.Get(ctx, key).Result()
	if err == redis.Nil {
		respondWithError(w, http.StatusNotFound, "Key not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := KeyValue{Key: key, Value: val}
	respondWithJSON(w, http.StatusOK, response)
}

// setValue handles POST requests to set a key-value pair in Redis.
func (a *App) setValue(w http.ResponseWriter, r *http.Request) {
	var kv KeyValue
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&kv); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	ctx := context.Background()
	err := a.RDB.Set(ctx, kv.Key, kv.Value, 0).Err()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, kv)
}

// main is the entry point of our application.
func main() {
	app := App{}
	app.initialize()
	app.run(":8080")
}
