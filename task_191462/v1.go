//go:build v1
// +build v1

package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
)

var client *mongo.Client
var dbcollection *mongo.Collection
var databaseName string
var inventoryCollection string
var usersCollection string
var jwtSecret string

type InventoryItem struct {
	ID          string  `json:"id,omitempty" bson:"_id,omitempty"`
	UserID      string  `json:"userID,omitempty" bson:"userID,omitempty"`
	ProductName string  `json:"productName" bson:"productName"`
	Units       int     `json:"units" bson:"units"`
	Price       float64 `json:"price" bson:"price"`
}

type User struct {
	ID       string `json:"id,omitempty" bson:"_id,omitempty"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

func initDB() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseURL := os.Getenv("MONGO_CONN_URL")
	clientOptions := options.Client().ApplyURI(databaseURL)
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")
}

func initCollection() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseName = os.Getenv("DATABASE_NAME")
	inventoryCollection = os.Getenv("INVENTORY_COLLECTION")
	usersCollection = os.Getenv("USERS_COLLECTION")
	jwtSecret = os.Getenv("JWT_SECRET")

	dbcollection = client.Database(databaseName).Collection(inventoryCollection)
}

func signUp(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
		})
		return
	}

	existingUser := User{}
	err := client.Database(databaseName).Collection(usersCollection).FindOne(context.Background(), bson.M{"username": user.Username}).Decode(&existingUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error checking username existence",
			})
			return
		}
	} else {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username already exists",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error hashing password",
		})
		return
	}

	user.Password = string(hashedPassword)

	_, err = client.Database(databaseName).Collection(usersCollection).InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error creating user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully!",
	})
}

func signIn(c *gin.Context) {
	// Implementation pending...
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation of token authentication middleware
	}
}

func getUserProducts(c *gin.Context) {
	userID, _ := c.Get("user")

	cursor, err := client.Database("Go-Gin-Inventory").Collection("inventory").Find(context.Background(), bson.M{"userID": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching products"})
		return
	}

	defer cursor.Close(context.Background())

	var products []InventoryItem
	if err := cursor.All(context.Background(), &products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func getProductById(c *gin.Context) {
	id := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	userID, _ := c.Get("user")

	var product InventoryItem
	err = dbcollection.FindOne(context.Background(), bson.M{"_id": objectId, "userID": userID}).Decode(&product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching product",
		})
		return
	}
	c.JSON(http.StatusOK, product)
}

func createProduct(c *gin.Context) {
	var product InventoryItem

	userID, _ := c.Get("user")
	product.UserID = userID.(string)

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	_, err := dbcollection.InsertOne(context.Background(), product)
	if err != nil {
		// Log the error for debugging, but don't return an error to the client.
		log.Println("Error inserting product:", err)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully!"})
}

func setupRoutes(r *gin.Engine) {
	r.POST("/signup", signUp)
	r.POST("/signin", signIn)

	authGroup := r.Group("/auth")
	authGroup.Use(authMiddleware())
	{
		authGroup.GET("/allProducts", getUserProducts)
		authGroup.GET("/products/:id", getProductById)
		authGroup.POST("/createProduct", createProduct)
	}
}

func main() {
	initDB()
	initCollection()
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"Authorization", "Content-Type"}
	r.Use(cors.New(config))

	setupRoutes(r)
	r.Run(":8080")
}
