package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subscriptions/internal/config"
	"subscriptions/internal/database"
	httpHandler "subscriptions/internal/delivery/http"
	"subscriptions/internal/repository"

	_ "subscriptions/docs"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Subscription Management API
// @version 1.0
// @description API Server for Subscription Tracking Application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	// 1. Load configuration
	log.Println("Initializing application...")
	cfg := config.MustLoad()

	// 2. Database connection string
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode)

	// 3. Run migrations
	if err := database.RunMigrations(dsn); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	// 4. Database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// 5. Initialize layers
	repo := repository.NewSubscriptionRepo(db)
	handler := httpHandler.SubscriptionHandler{Repo: repo}

	// 6. Routes configuration
	r := mux.NewRouter()

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API Endpoints
	r.HandleFunc("/subscriptions", handler.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions/stats/{user_id}/{from}/{to}/{sub_name}", handler.GetStats).Methods("GET")
	r.HandleFunc("/subscriptions/by-id/{user_id}/{sub_id}", handler.ReadSubByID).Methods("GET")
	r.HandleFunc("/subscriptions/list/{user_id}", handler.GetSubList).Methods("GET")
	r.HandleFunc("/subscriptions/{user_id}/{sub_id}", handler.DeleteSubByID).Methods("DELETE")
	r.HandleFunc("/subscriptions/{user_id}/{sub_id}", handler.UpdateSubByID).Methods("PUT")

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// 7. Graceful Shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server started successfully on port %s", cfg.Port)
		log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
