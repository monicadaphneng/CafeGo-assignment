package main

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var database *sql.DB

func initDatabase() {
	var err error
	database, err = sql.Open("sqlite", "cafego.db")
	if err != nil {
		log.Fatal(err)
	}

	queries := []string{
		"CREATE TABLE IF NOT EXISTS cgo_user (username TEXT, password TEXT)",
		"CREATE TABLE IF NOT EXISTS cgo_product (name TEXT, price INTEGER, description TEXT)",
		"CREATE TABLE IF NOT EXISTS cgo_session (token TEXT, user_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS cgo_cart_item (product_id INTEGER, quantity INTEGER, user_id INTEGER)",
		"CREATE TABLE IF NOT EXISTS cgo_transaction (user_id INTEGER, created_at TEXT)",
		"CREATE TABLE IF NOT EXISTS cgo_line_item (transaction_id INTEGER, product_id INTEGER, quantity INTEGER)",
	}

	for _, q := range queries {
		if _, err := database.Exec(q); err != nil {
			log.Fatal(err)
		}
	}
}

type User struct {
	Id       int
	Username string
	Password string
}

type Product struct {
	Id          int
	Name        string
	Price       int
	Description string
}

type CartItem struct {
	Id          int
	UserId      int
	ProductId   int
	Quantity    int
	ProductName string
}

type Transaction struct {
	ID        int
	UserId    int
	CreatedAt string
	Items     []CartItem
}

// Users
func getUsers() []User {
	rows, _ := database.Query("SELECT rowid, username, password FROM cgo_user")
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.Id, &u.Username, &u.Password)
		users = append(users, u)
	}
	return users
}

func getUserByUsername(username string) (User, bool) {
	row := database.QueryRow("SELECT rowid, username, password FROM cgo_user WHERE username = ?", username)
	var u User
	err := row.Scan(&u.Id, &u.Username, &u.Password)
	if err != nil {
		return User{}, false
	}
	return u, true
}

func getUserFromSessionToken(token string) User {
	var user User
	row := database.QueryRow(`
		SELECT cgo_user.rowid, username, password
		FROM cgo_session
		LEFT JOIN cgo_user ON cgo_session.user_id = cgo_user.rowid
		WHERE token = ?
	`, token)
	row.Scan(&user.Id, &user.Username, &user.Password)
	return user
}

func setSession(token string, user User) {
	database.Exec("INSERT INTO cgo_session(token, user_id) VALUES (?, ?)", token, user.Id)
}

// Products
func getProducts() []Product {
	rows, _ := database.Query("SELECT rowid, name, price, description FROM cgo_product GROUP BY name")
	defer rows.Close()
	var products []Product
	for rows.Next() {
		var p Product
		rows.Scan(&p.Id, &p.Name, &p.Price, &p.Description)
		products = append(products, p)
	}
	return products
}

func getProductById(id int) (Product, bool) {
	row := database.QueryRow("SELECT rowid, name, price, description FROM cgo_product WHERE rowid = ?", id)
	var p Product
	err := row.Scan(&p.Id, &p.Name, &p.Price, &p.Description)
	if err != nil {
		return Product{}, false
	}
	return p, true
}

// Cart
func getCartItemsByUser(user User) []CartItem {
	rows, _ := database.Query(`
		SELECT cgo_cart_item.rowid, cgo_cart_item.user_id, cgo_cart_item.product_id, SUM(cgo_cart_item.quantity), cgo_product.name
		FROM cgo_cart_item
		LEFT JOIN cgo_product ON cgo_cart_item.product_id = cgo_product.rowid
		WHERE cgo_cart_item.user_id = ?
		GROUP BY cgo_cart_item.product_id
	`, user.Id)
	defer rows.Close()

	var result []CartItem
	for rows.Next() {
		var ci CartItem
		rows.Scan(&ci.Id, &ci.UserId, &ci.ProductId, &ci.Quantity, &ci.ProductName)
		result = append(result, ci)
	}
	return result
}

func addItemToCart(user User, product Product, quantity int) {
	database.Exec("INSERT INTO cgo_cart_item(product_id, quantity, user_id) VALUES (?, ?, ?)", product.Id, quantity, user.Id)
}

// Transactions
func checkoutItemsForUser(user User) {
	cartItems := getCartItemsByUser(user)
	if len(cartItems) == 0 {
		return
	}

	now := time.Now().UTC()
	res, err := database.Exec("INSERT INTO cgo_transaction(user_id, created_at) VALUES (?, ?)", user.Id, now)
	if err != nil {
		log.Fatal(err)
	}

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	for _, ci := range cartItems {
		database.Exec("INSERT INTO cgo_line_item(transaction_id, product_id, quantity) VALUES (?, ?, ?)", lastInsertId, ci.ProductId, ci.Quantity)
		database.Exec("DELETE FROM cgo_cart_item WHERE rowid = ?", ci.Id)
	}
}

func getTransactionsByUser(user User) []Transaction {
	rows, _ := database.Query("SELECT rowid, user_id, created_at FROM cgo_transaction WHERE user_id = ? ORDER BY created_at DESC", user.Id)
	defer rows.Close()
	var txs []Transaction
	for rows.Next() {
		var t Transaction
		rows.Scan(&t.ID, &t.UserId, &t.CreatedAt)
		t.Items = getLineItemsByTransaction(t.ID)
		txs = append(txs, t)
	}
	return txs
}

func getLineItemsByTransaction(transactionId int) []CartItem {
	rows, _ := database.Query(`
		SELECT cgo_line_item.rowid, cgo_line_item.product_id, cgo_line_item.quantity, cgo_product.name
		FROM cgo_line_item
		LEFT JOIN cgo_product ON cgo_line_item.product_id = cgo_product.rowid
		WHERE cgo_line_item.transaction_id = ?
	`, transactionId)
	defer rows.Close()

	var items []CartItem
	for rows.Next() {
		var ci CartItem
		rows.Scan(&ci.Id, &ci.ProductId, &ci.Quantity, &ci.ProductName)
		items = append(items, ci)
	}
	return items
}

