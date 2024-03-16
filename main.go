package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Bill struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	Price       float32 `json:"price"`
	PaymentDate string  `json:"payment_date"`
	Payed       bool    `json:"payed"`
}

func main() {
	initDB()

	http.HandleFunc("/bills_by_category", handleBills)
	http.HandleFunc("/bills/add", handleAddBill)
	http.HandleFunc("/bills", handleAllBills)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	cfg := "test:password@tcp(127.0.0.1:3306)/test"
	var err error
	db, err = sql.Open("mysql", cfg)
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected to the database!")
}

func handleBills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	category := r.URL.Query().Get("category")

	bills, err := getBillsByCategory(category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(bills)
}

func handleAllBills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bills, err := getAllBills()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(bills)
}

func handleAddBill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newBill Bill
	err := json.NewDecoder(r.Body).Decode(&newBill)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := addBill(newBill)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		ID int64 `json:"id"`
	}{
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getBillsByCategory(category string) ([]Bill, error) {
	var bills []Bill

	rows, err := db.Query("SELECT * FROM Bills WHERE Category = ?", category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bill Bill
		if err := rows.Scan(&bill.ID, &bill.Title, &bill.Category, &bill.Price, &bill.PaymentDate, &bill.Payed); err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bills, nil
}

func addBill(b Bill) (int64, error) {
	result, err := db.Exec("INSERT INTO Bills (Title, Category, Price, PaymentDate, Payed) VALUES (?, ?, ?, ?, ?)", b.Title, b.Category, b.Price, b.PaymentDate, b.Payed)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func getAllBills() ([]Bill, error) {
	var bills []Bill

	rows, err := db.Query("SELECT * FROM Bills")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bill Bill
		if err := rows.Scan(&bill.ID, &bill.Title, &bill.Category, &bill.Price, &bill.PaymentDate, &bill.Payed); err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bills, nil
}
