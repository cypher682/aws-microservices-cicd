package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Product struct {
	ProductID   string  `json:"productId" dynamodbav:"productId"`
	Name        string  `json:"name" dynamodbav:"name"`
	Description string  `json:"description" dynamodbav:"description"`
	Price       float64 `json:"price" dynamodbav:"price"`
	Category    string  `json:"category" dynamodbav:"category"`
	Stock       int     `json:"stock" dynamodbav:"stock"`
	CreatedAt   string  `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt   string  `json:"updatedAt" dynamodbav:"updatedAt"`
}

var (
	db        *dynamodb.DynamoDB
	tableName string

	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_service_requests_total",
			Help: "Total requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "product_service_request_duration_seconds",
			Help: "Request duration",
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(getEnv("AWS_REGION", "us-east-1")),
	}))
	db = dynamodb.New(sess)
	tableName = getEnv("DYNAMODB_PRODUCTS_TABLE", "aws-microservices-cicd-products")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		
		status := c.Writer.Status()
		requestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
		requestCount.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Inc()
	}
}

func main() {
	r := gin.Default()
	r.Use(prometheusMiddleware())

	r.GET("/health", healthHandler)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	
	r.POST("/", createProduct)
	r.GET("/", listProducts)
	r.GET("/:id", getProduct)
	r.PUT("/:id", updateProduct)
	r.DELETE("/:id", deleteProduct)

	port := getEnv("PORT", "8080")
	log.Printf("Product service listening on port %s", port)
	r.Run(":" + port)
}

func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "healthy", "service": "product-service"})
}

func createProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	product.ProductID = uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	product.CreatedAt = now
	product.UpdatedAt = now

	av, err := dynamodbattribute.MarshalMap(product)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal product"})
		return
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	if _, err := db.PutItem(input); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, product)
}

func getProduct(c *gin.Context) {
	productID := c.Param("id")

	result, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"productId": {S: aws.String(productID)},
		},
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if result.Item == nil {
		c.JSON(404, gin.H{"error": "Product not found"})
		return
	}

	var product Product
	if err := dynamodbattribute.UnmarshalMap(result.Item, &product); err != nil {
		c.JSON(500, gin.H{"error": "Failed to unmarshal product"})
		return
	}

	c.JSON(200, product)
}

func listProducts(c *gin.Context) {
	result, err := db.Scan(&dynamodb.ScanInput{
		TableName: aws.String(tableName),
		Limit:     aws.Int64(10),
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var products []Product
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &products); err != nil {
		c.JSON(500, gin.H{"error": "Failed to unmarshal products"})
		return
	}

	c.JSON(200, gin.H{"products": products, "count": len(products)})
}

func updateProduct(c *gin.Context) {
	productID := c.Param("id")

	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	product.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"productId": {S: aws.String(productID)},
		},
		UpdateExpression: aws.String("SET #name = :name, description = :desc, price = :price, category = :cat, stock = :stock, updatedAt = :updated"),
		ExpressionAttributeNames: map[string]*string{
			"#name": aws.String("name"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name":    {S: aws.String(product.Name)},
			":desc":    {S: aws.String(product.Description)},
			":price":   {N: aws.String(fmt.Sprintf("%f", product.Price))},
			":cat":     {S: aws.String(product.Category)},
			":stock":   {N: aws.String(fmt.Sprintf("%d", product.Stock))},
			":updated": {S: aws.String(product.UpdatedAt)},
		},
		ConditionExpression: aws.String("attribute_exists(productId)"),
		ReturnValues:        aws.String("ALL_NEW"),
	}

	result, err := db.UpdateItem(input)
	if err != nil {
		c.JSON(404, gin.H{"error": "Product not found"})
		return
	}

	var updatedProduct Product
	if err := dynamodbattribute.UnmarshalMap(result.Attributes, &updatedProduct); err != nil {
		c.JSON(500, gin.H{"error": "Failed to unmarshal product"})
		return
	}

	c.JSON(200, updatedProduct)
}

func deleteProduct(c *gin.Context) {
	productID := c.Param("id")

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"productId": {S: aws.String(productID)},
		},
		ConditionExpression: aws.String("attribute_exists(productId)"),
	}

	if _, err := db.DeleteItem(input); err != nil {
		c.JSON(404, gin.H{"error": "Product not found"})
		return
	}

	c.Status(204)
}
