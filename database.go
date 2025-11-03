package main

import (
    "database/sql"
    _ "modernc.org/sqlite"
    "log"
)

var database *sql.DB

func initDB() {
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

func getUserFromSessionToken(token string) User {
    var user User
    row := database.QueryRow("SELECT cgo_user.rowid, username, password FROM cgo_session LEFT JOIN cgo_user ON cgo_session.user_id = cgo_user.rowid WHERE token = ?", token)
    row.Scan(&user.Id, &user.Username, &user.Password)
    return user
}

func setSession(token string, user User) {
    database.Exec("INSERT INTO cgo_session(token, user_id) VALUES (?, ?)", token, user.Id)
}

func getProducts() []Product {
    rows, _ := database.Query("SELECT rowid, name, price, description FROM cgo_product")
    defer rows.Close()
    var products []Product
    for rows.Next() {
        var p Product
        rows.Scan(&p.Id, &p.Name, &p.Price, &p.Description)
        products = append(products, p)
    }
    return products
}

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

