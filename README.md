# Subscription Management Service

A high-performance backend service written in Go for managing user subscriptions and tracking expenses.

## 🚀 Features

- Full CRUD: Create, Read, Update, and Delete subscriptions.
- Spending Statistics: Calculate total expenses for specific services within a date range.
- Database Migrations: Automatic schema updates on startup.
- Dockerized: Easy to deploy using Docker and Docker Compose.
- API Documentation: Integrated Swagger UI (OpenAPI 3.0).

## 🛠 Tech Stack

- Language: Go (Golang) 1.26.2
- Database: PostgreSQL 15
- Router: Gorilla Mux
- Containerization: Docker / Docker Compose
- Documentation: Swag (Swagger)

## 📋 Prerequisites

- Docker and Docker Compose installed.
- (Optional) Go 1.26.2+ if running locally.

## ⚙️ Installation & Setup

1. Clone the repository:
   bash
   git clone [https://github.com/your-username/subscription-service.git](https://github.com/your-username/subscription-service.git)
   cd subscription-service2. Configure Environment Variables
   Copy the example environment file and fill in your database credentials:
   Bash
   cp .env.example .env
   - Note: For Docker setup, keep DB_HOST=db.

2. Run with Docker Compose
   Bash
   docker-compose up --build
   The service will be available at:
   - http://localhost:8080

## 📖 API Documentation

- Once the service is running, Swagger documentation is available at:

http://localhost:8080/swagger/index.html

## 🏗 Project Structure

- cmd/app - Application entry point
- internal/models - Data structures and entities
- internal/repository - Database layer (PostgreSQL)
- internal/httpHandler - HTTP handlers and routing
- internal/database/migrations - SQL migration files

## 🐳 Docker Support

The project is fully containerized and can be started with a single command using Docker Compose.
Included services:

- Go application
- PostgreSQL database

## 📌 Notes

1. Environment variables are managed through .env.
2. Database migrations run automatically on application startup.
3. Swagger documentation is generated using swag.

## 📄 License

This project is created for educational and portfolio purposes.
