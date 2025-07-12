package examples

import (
	"fmt"
	"log"
	"os"
	"time"
	
	"goodoo/database"
	"goodoo/models"
	"gorm.io/gorm"
)

// ExampleDatabaseUsage demonstrates how to use the database layer
func ExampleDatabaseUsage() {
	// Initialize the database system
	opts := database.DefaultInitOptions()
	if err := database.Initialize(opts); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Configure database connection
	config := database.DefaultConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.User = "postgres"
	config.Password = "password"
	config.Database = "goodoo_dev"
	config.SSLMode = "disable"
	
	// Or load from environment variables
	config.LoadFromEnv()
	
	// Register the database
	if err := database.RegisterDatabase("goodoo_dev", config); err != nil {
		log.Fatalf("Failed to register database: %v", err)
	}
	
	// Get a database connection
	conn, err := database.GetDatabaseConnection("goodoo_dev")
	if err != nil {
		log.Fatalf("Failed to get connection: %v", err)
	}
	defer conn.Close()
	
	// Create environment for ORM operations
	env, err := models.NewEnvironmentForDB("goodoo_dev", 1) // user ID 1
	if err != nil {
		log.Fatalf("Failed to create environment: %v", err)
	}
	
	// Auto-migrate models
	db := env.GetDB()
	if err := db.AutoMigrate(
		&models.User{},
		&models.Partner{},
		&models.Product{},
		&models.ProductCategory{},
		&models.SaleOrder{},
		&models.SaleOrderLine{},
	); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
	
	// Example CRUD operations
	exampleCRUDOperations(env)
	
	// Example transaction usage
	exampleTransactionUsage(conn)
	
	// Example relationship operations
	exampleRelationshipOperations(env)
	
	// Cleanup
	database.Cleanup()
}

func exampleCRUDOperations(env *models.Environment) {
	fmt.Println("=== CRUD Operations Example ===")
	
	// Create users
	userRS := models.Model(env, models.User{})
	users, err := userRS.Create([]models.User{
		{Name: "John Doe", Email: "john@example.com", Login: "john", Active: true},
		{Name: "Jane Smith", Email: "jane@example.com", Login: "jane", Active: true},
	})
	if err != nil {
		log.Printf("Error creating users: %v", err)
		return
	}
	fmt.Printf("Created %d users\n", len(users.Records))
	
	// Search for users
	domain := models.Domain{
		[]interface{}{"active", "=", true},
		[]interface{}{"email", "like", "%@example.com"},
	}
	
	foundUsers, err := userRS.Search(domain, 0, 10, "name")
	if err != nil {
		log.Printf("Error searching users: %v", err)
		return
	}
	fmt.Printf("Found %d users\n", len(foundUsers.Records))
	
	// Read specific fields
	userData, err := foundUsers.Read([]string{"name", "email"})
	if err != nil {
		log.Printf("Error reading user data: %v", err)
		return
	}
	fmt.Printf("Read %d user records\n", len(userData))
	
	// Update users
	err = foundUsers.Write(map[string]interface{}{
		"active": true,
	})
	if err != nil {
		log.Printf("Error updating users: %v", err)
		return
	}
	fmt.Println("Updated users successfully")
	
	// Count users
	count, err := userRS.Count(domain)
	if err != nil {
		log.Printf("Error counting users: %v", err)
		return
	}
	fmt.Printf("Total active users: %d\n", count)
}

func exampleTransactionUsage(conn *database.Connection) {
	fmt.Println("\n=== Transaction Usage Example ===")
	
	// Using GORM transactions
	err := conn.Transaction(func(tx *gorm.DB) error {
		// Create a partner
		partner := models.Partner{
			Name:  "Acme Corp",
			Email: "info@acme.com",
		}
		if err := tx.Create(&partner).Error; err != nil {
			return err
		}
		
		// Create users for this partner
		users := []models.User{
			{Name: "John Acme", Email: "john@acme.com", Login: "john_acme", PartnerID: &partner.ID},
			{Name: "Jane Acme", Email: "jane@acme.com", Login: "jane_acme", PartnerID: &partner.ID},
		}
		if err := tx.Create(&users).Error; err != nil {
			return err
		}
		
		fmt.Printf("Created partner %s with %d users in transaction\n", partner.Name, len(users))
		return nil
	})
	
	if err != nil {
		log.Printf("Transaction failed: %v", err)
	} else {
		fmt.Println("Transaction completed successfully")
	}
}

func exampleRelationshipOperations(env *models.Environment) {
	fmt.Println("\n=== Relationship Operations Example ===")
	
	// Create a product category
	categoryRS := models.Model(env, models.ProductCategory{})
	categories, err := categoryRS.Create([]models.ProductCategory{
		{Name: "Electronics", CompleteName: "Electronics"},
	})
	if err != nil {
		log.Printf("Error creating category: %v", err)
		return
	}
	category := categories.Records[0]
	
	// Create products in this category
	productRS := models.Model(env, models.Product{})
	products, err := productRS.Create([]models.Product{
		{Name: "Laptop", DefaultCode: "LAP001", ListPrice: 999.99, CategoryID: &category.ID},
		{Name: "Mouse", DefaultCode: "MOU001", ListPrice: 29.99, CategoryID: &category.ID},
	})
	if err != nil {
		log.Printf("Error creating products: %v", err)
		return
	}
	fmt.Printf("Created %d products in category %s\n", len(products.Records), category.Name)
	
	// Query products with category information using GORM preloading
	var productsWithCategory []models.Product
	err = env.GetDB().Preload("Category").Find(&productsWithCategory).Error
	if err != nil {
		log.Printf("Error loading products with category: %v", err)
		return
	}
	
	for _, product := range productsWithCategory {
		if product.Category != nil {
			fmt.Printf("Product: %s, Category: %s\n", product.Name, product.Category.Name)
		}
	}
}

// ExampleEnvironmentVariables shows how to use environment variables for configuration
func ExampleEnvironmentVariables() {
	// Set environment variables (normally done in shell or docker)
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("DB_NAME", "goodoo_dev")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("DB_MAXCONN", "32")
	os.Setenv("GOODOO_PGAPPNAME", "goodoo-{pid}")
	
	// Quick setup using environment variables
	err := database.QuickSetup("goodoo_dev",
		&models.User{},
		&models.Partner{},
		&models.Product{},
		&models.ProductCategory{},
		&models.SaleOrder{},
		&models.SaleOrderLine{},
	)
	
	if err != nil {
		log.Fatalf("Quick setup failed: %v", err)
	}
	
	fmt.Println("Database setup completed using environment variables")
	
	// Health check
	results := database.HealthCheck()
	for dbName, err := range results {
		if err != nil {
			fmt.Printf("Database %s: ERROR - %v\n", dbName, err)
		} else {
			fmt.Printf("Database %s: OK\n", dbName)
		}
	}
}

// ExampleConnectionPooling demonstrates connection pool usage
func ExampleConnectionPooling() {
	fmt.Println("\n=== Connection Pooling Example ===")
	
	// Get pool statistics
	stats := database.Stats()
	fmt.Printf("Pool stats: %s\n", stats.String())
	
	// Get registry statistics
	registry := database.GetRegistry()
	registryStats := registry.Stats()
	fmt.Printf("Registry stats: %s\n", registryStats.String())
	
	// Connect to database using URI
	conn, err := database.Connect("postgresql://user:pass@localhost/dbname", true)
	if err != nil {
		log.Printf("URI connection failed: %v", err)
	} else {
		fmt.Println("Connected using URI")
		conn.Close()
	}
	
	// Cleanup inactive connections
	registry.CleanupInactive(30 * time.Minute)
	fmt.Println("Cleaned up inactive connections")
}