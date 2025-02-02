package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique" json:"username"`
	Email     string    `gorm:"unique" json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

var db *gorm.DB

func main() {
	connectDB()
	r := mux.NewRouter()

	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/{username}", getUser).Methods("GET")

	http.ListenAndServe(":8080", r)
}

func connectDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"))

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.CreatedAt = time.Now()
	result := db.Create(&user)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var user User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Add HTTP handlers for createUser and getUser here
