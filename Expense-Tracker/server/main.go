package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Expense struct {
	ID          string    `json:"id,omitempty"`
	Title       string    `json:"title" bson:"title"`
	Amount      float64   `json:"amount" bson:"amount"`
	Category    string    `json:"category" bson:"category"`
	Date        time.Time `json:"date" bson:"date"`
	Description string    `json:"description" bson:"description"`
}

var (
	client     *mongo.Client
	collection *mongo.Collection
	ctx        = context.Background()
)

func main() {
	_ = godotenv.Load()
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not set in .env")
	}
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	collection = client.Database("Expense_Tracker").Collection("expense_data")

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Expense tracker is running")
	})
	r.POST("/expense", createExpense)
	r.GET("/expense", getExpense)
	r.GET("/expense/:id", getExpenseByID)
	r.PUT("/expense/:id", updateExpense)
	r.DELETE("/expense/:id", deleteExpense)
	r.GET("/categories", getCategories)
	r.Run(":8080")
}

func createExpense(c *gin.Context) {
	var newExpense Expense
	if err := c.BindJSON(&newExpense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if newExpense.Title == "" || newExpense.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and amount are required"})
		return
	}
	newExpense.Date = time.Now()

	doc := bson.M{
		"title":       newExpense.Title,
		"amount":      newExpense.Amount,
		"category":    newExpense.Category,
		"date":        newExpense.Date,
		"description": newExpense.Description,
	}

	res, err := collection.InsertOne(ctx, doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert expense"})
		return
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inserted ID"})
		return
	}
	newExpense.ID = oid.Hex()
	c.JSON(http.StatusCreated, newExpense)
}

func getExpense(c *gin.Context) {
	category := c.Query("category")
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expenses"})
		return
	}
	defer cur.Close(ctx)
	var list []Expense
	for cur.Next(ctx) {
		var e Expense
		var raw bson.M
		if err := cur.Decode(&raw); err == nil {
			e.ID = raw["_id"].(primitive.ObjectID).Hex()
			e.Title = raw["title"].(string)
			e.Amount = raw["amount"].(float64)
			e.Category = raw["category"].(string)
			e.Description = raw["description"].(string)
			e.Date = raw["date"].(primitive.DateTime).Time()
			list = append(list, e)
		}
	}
	c.JSON(http.StatusOK, list)
}

func getExpenseByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var e Expense
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&e)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expense"})
		return
	}
	e.ID = objID.Hex()
	c.JSON(http.StatusOK, e)
}

func updateExpense(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var updated Expense
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if updated.Title == "" || updated.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and amount are required"})
		return
	}
	update := bson.M{
		"$set": bson.M{
			"title":       updated.Title,
			"amount":      updated.Amount,
			"category":    updated.Category,
			"description": updated.Description,
		},
	}
	res, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update expense"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Expense updated"})
}

func deleteExpense(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	res, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expense"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func getCategories(c *gin.Context) {
	catCur, err := collection.Distinct(ctx, "category", bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	var cats []string
	for _, v := range catCur {
		if s, ok := v.(string); ok && s != "" {
			cats = append(cats, s)
		}
	}
	c.JSON(http.StatusOK, cats)
}
