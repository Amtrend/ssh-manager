package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"ssh_manager/internal/encryption"
	"ssh_manager/internal/handlers"
	"ssh_manager/internal/middleware"
	"ssh_manager/internal/repository"
	"ssh_manager/internal/services"
	"ssh_manager/internal/utils"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func main() {
	// Logger initialization and page template cache
	utils.InitLogger()
	utils.InitTemplates()

	// Loading environment variables
	_ = godotenv.Load()

	// Database initialization
	var db *sql.DB
	var err error
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	if dbType == "postgres" {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
		db, err = sql.Open("postgres", dsn)
	} else {
		// Get the database name from ENV or set the default
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "ssh_manager"
		}

		// Create a path. Add the .db extension if it's not in the config.
		dbPath := fmt.Sprintf("./data/%s.db", dbName)

		os.MkdirAll("./data", os.ModePerm)
		db, err = sql.Open("sqlite", dbPath)
		log.Printf("Using SQLite database file: %s", dbPath)
	}

	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	defer db.Close()

	// Setting up a pool for sql.DB
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Checking the connection to the database (Ping)
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot reach database: %v", err)
	}

	// Rolling out tables (Migrations)
	if err := repository.InitDB(db, dbType); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Create an initial admin
	adminUser := utils.GetEnv("INITIAL_ADMIN_USER", "admin")
	adminPass := utils.GetEnv("INITIAL_ADMIN_PASSWORD", "admin")
	repository.EnsureAdminUser(db, dbType, adminUser, adminPass)

	// Read the settings for the cleaner
	cleanupInterval := utils.GetDurationEnv("CLEANUP_INTERVAL", "2m")
	sessionTimeout := utils.GetDurationEnv("SESSION_TIMEOUT", "10m")

	// Sessions
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		Secure:   false, // В проде на HTTPS ставь true
		SameSite: http.SameSiteLaxMode,
	}

	// Encryption
	hexKey := os.Getenv("ENCRYPTION_KEY")
	encryptionKey, err := hex.DecodeString(hexKey)
	if err != nil || len(encryptionKey) != 32 {
		log.Fatalf("Encryption key error: %v", err)
	}
	encryption.SetEncryptionKey(encryptionKey)

	// Services and repositories
	uRepo := &repository.UserRepository{DB: db}
	hRepo := &repository.HostRepository{DB: db}
	kRepo := &repository.KeyRepository{DB: db}
	sshService := services.NewSSHService(hRepo, kRepo, cleanupInterval, sessionTimeout)

	handler := &handlers.Handlers{
		UserRepo: uRepo, KeyRepo: kRepo, HostRepo: hRepo,
		Store: store, SSHService: sshService,
	}

	authMiddleware := &middleware.Middleware{Store: store}

	// Router
	r := SetupRoutes(handler, authMiddleware, store)

	port := utils.GetEnv("PORT", "8080")
	addr := ":" + port
	log.Printf("Server started on %s (Mode: %s)", addr, dbType)
	// Starting the server
	log.Fatal(http.ListenAndServe(addr, r))
}
