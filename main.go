package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func main() {
	initDatabase()

	// Seed example data if empty
	if len(getProducts()) == 0 {
		database.Exec("INSERT INTO cgo_product(name, price, description) VALUES (?, ?, ?)", "Americano", 100, "Hot brewed coffee")
		database.Exec("INSERT INTO cgo_product(name, price, description) VALUES (?, ?, ?)", "Cappuccino", 110, "Espresso with milk foam")
		database.Exec("INSERT INTO cgo_product(name, price, description) VALUES (?, ?, ?)", "Espresso", 90, "Strong black coffee")
		database.Exec("INSERT INTO cgo_product(name, price, description) VALUES (?, ?, ?)", "Macchiato", 120, "Espresso with a dash of milk")
	}

	// Seed a user if empty
	if _, exists := getUserByUsername("melinoe"); !exists {
		database.Exec("INSERT INTO cgo_user(username, password) VALUES (?, ?)", "melinoe", "1234")
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/product/", productHandler)
	http.HandleFunc("/cart/", cartHandler)
	http.HandleFunc("/transactions/", transactionsHandler)

	fmt.Println("Server running at http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	products := getProducts()
	user := User{Username: "melinoe"} // Default for demo

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, struct {
		Username string
		Products []Product
	}{
		Username: user.Username,
		Products: products,
	})
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/product/"):]
	id, _ := strconv.Atoi(idStr)
	product, ok := getProductById(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	user := User{Id: 1, Username: "melinoe"} // default user

	if r.Method == "POST" {
		qty, _ := strconv.Atoi(r.FormValue("quantity"))
		if qty < 1 {
			qty = 1
		}
		addItemToCart(user, product, qty)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/product.html"))
	tmpl.Execute(w, product)
}

func cartHandler(w http.ResponseWriter, r *http.Request) {
	user := User{Id: 1, Username: "melinoe"} // default

	if r.Method == "POST" {
		checkoutItemsForUser(user)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cartItems := getCartItemsByUser(user)
	tmpl := template.Must(template.ParseFiles("templates/cart.html"))
	tmpl.Execute(w, struct {
		User      User
		CartItems []CartItem
	}{
		User:      user,
		CartItems: cartItems,
	})
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	user := User{Id: 1, Username: "melinoe"} // default
	txs := getTransactionsByUser(user)

	tmpl := template.Must(template.ParseFiles("templates/transactions.html"))
	tmpl.Execute(w, struct {
		User         User
		Transactions []Transaction
	}{
		User:         user,
		Transactions: txs,
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		user, ok := getUserByUsername(username)
		if ok && user.Password == password {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	tmpl.Execute(w, nil)
}

