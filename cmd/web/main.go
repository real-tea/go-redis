package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Transaction represents a single financial transaction.
type Transaction struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
}

// In-memory "database" for our transactions.
var transactions = []Transaction{}
var nextID = 1

// getTransactionsHandler sends the list of all transactions as JSON.
func getTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// addTransactionHandler adds a new transaction from a JSON payload.
func addTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var t Transaction
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Assign a new ID and the current date
	t.ID = nextID
	t.Date = time.Now()
	nextID++

	transactions = append(transactions, t)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func main() {
	// Create a new router
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTransactionsHandler(w, r)
		case http.MethodPost:
			addTransactionHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Create a file server to serve the static frontend files (HTML, CSS, JS).
	// This tells Go to look in the './ui/static/' directory for files.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// We use http.StripPrefix to remove the '/static/' part of the URL path,
	// so a request to '/static/app.js' correctly looks for 'app.js' in the directory.
	mux.Handle("/", fileServer)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
