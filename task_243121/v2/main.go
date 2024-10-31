package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
)

// Model: Product struct
type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
}

// ViewModel: ProductViewModel struct
type ProductViewModel struct {
	Product   Product
	IsEditing bool
}

// Database connection details
const (
	DB_HOST = "localhost"
	DB_PORT = "3306"
	DB_USER = "your_db_user"
	DB_PASS = "your_db_password"
	DB_NAME = "your_db_name"
)

// Database connection
var db *sql.DB

//func init() {
//	var err error
//	dbSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", DB_USER, DB_PASS, DB_HOST, DB_PORT, DB_NAME)
//	db, err = sql.Open("mysql", dbSource)
//	if err != nil {
//		panic(err.Error())
//	}
//
//	// Create the table if it doesn't exist
//	createTableSQL := `
//		CREATE TABLE IF NOT EXISTS products (
//		  id INT AUTO_INCREMENT PRIMARY KEY,
//		  name VARCHAR(255) NOT NULL,
//		  description TEXT,
//		  price DECIMAL(10,2) NOT NULL
//		);
//	`
//	_, err = db.Exec(createTableSQL)
//	if err != nil {
//		panic(err.Error())
//	}
//}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/delete", deleteHandler)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- Handlers ---

func indexHandler(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, products)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/create.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := getProductByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	viewModel := ProductViewModel{
		Product:   product,
		IsEditing: true,
	}

	tmpl, err := template.ParseFiles("templates/create.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, viewModel)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	priceStr := r.FormValue("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	err = updateProduct(id, name, description, price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = deleteProduct(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- Database operations ---

func getProducts() ([]Product, error) {
	rows, err := db.Query("SELECT * FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func getProductByID(id int) (Product, error) {
	var p Product
	err := db.QueryRow("SELECT * FROM products WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Description, &p.Price)
	if err != nil {
		return Product{}, err
	}
	return p, nil
}

func updateProduct(id int, name, description string, price float64) error {
	_, err := db.Exec("UPDATE products SET name = ?, description = ?, price = ? WHERE id = ?", name, description, price, id)
	return err
}

func deleteProduct(id int) error {
	_, err := db.Exec("DELETE FROM products WHERE id = ?", id)
	return err
}
