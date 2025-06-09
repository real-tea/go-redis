package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"golang.org/x/crypto/bcrypt"
)

// --- Structs ---

type User struct {
	ID       int
	Username string
	Password string // This will be the hashed password
}

type Transaction struct {
	ID          int       `json:"id"`
	UserID      int       `json:"-"` // Foreign key to User, ignored in JSON output
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
}

// Global database connection pool
var db *sql.DB

// --- Database Initialization ---

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./moneytracker.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create users table
	createUserTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"username" TEXT NOT NULL UNIQUE,
		"password" TEXT NOT NULL
	  );`
	_, err = db.Exec(createUserTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Create transactions table
	createTransactionTableSQL := `CREATE TABLE IF NOT EXISTS transactions (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"user_id" INTEGER NOT NULL,
		"description" TEXT,
		"amount" REAL NOT NULL,
		"date" DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`
	_, err = db.Exec(createTransactionTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database initialized successfully")
}

// --- Handlers ---

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error, unable to create your account.", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, string(hashedPassword))
	if err != nil {
		// This likely means the username is already taken
		http.Error(w, "Username already exists.", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    // For this simple example, we'll use basic auth for login.
    // A real app would use session tokens (e.g., JWT).
    username, password, ok := r.BasicAuth()
    if !ok {
        w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var storedPassword string
    var userID int
    err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&userID, &storedPassword)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Invalid username or password", http.StatusUnauthorized)
            return
        }
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
    if err != nil {
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }
    
    // In a real app, you would issue a token here.
    // For simplicity, successful basic auth will be our "token".
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}


func transactionsHandler(w http.ResponseWriter, r *http.Request) {
    username, _, ok := r.BasicAuth()
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var userID int
    err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    switch r.Method {
    case http.MethodGet:
        getTransactions(w, r, userID)
    case http.MethodPost:
        addTransaction(w, r, userID)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func getTransactions(w http.ResponseWriter, r *http.Request, userID int) {
    rows, err := db.Query("SELECT id, description, amount, date FROM transactions WHERE user_id = ?", userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var transactions []Transaction
    for rows.Next() {
        var t Transaction
        if err := rows.Scan(&t.ID, &t.Description, &t.Amount, &t.Date); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        transactions = append(transactions, t)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(transactions)
}

func addTransaction(w http.ResponseWriter, r *http.Request, userID int) {
    var t Transaction
    if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    t.UserID = userID
    t.Date = time.Now()

    result, err := db.Exec(
        "INSERT INTO transactions (user_id, description, amount, date) VALUES (?, ?, ?, ?)",
        t.UserID, t.Description, t.Amount, t.Date,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    id, _ := result.LastInsertId()
    t.ID = int(id)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(t)
}


// --- Main Function ---
func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/register", registerHandler)
	mux.HandleFunc("/api/login", loginHandler)
    mux.HandleFunc("/api/transactions", transactionsHandler)

	// Serve frontend files
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/", fileServer)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
