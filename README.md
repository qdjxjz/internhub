# InternHub 🚀  
A Go-based Microservices Backend for Internship Recruitment Platform  

## 📖 Overview

InternHub is a microservices-based backend system built with Go, designed to simulate a real-world internship recruitment platform.

This project demonstrates:

- Microservices architecture
- API Gateway pattern
- JWT-based authentication
- Stateless authorization
- PostgreSQL integration
- GORM ORM usage
- Password hashing with bcrypt
- Go multi-module workspace (`go.work`)

---

## 🏗 Architecture

Client  
   ↓  
API Gateway (JWT Validation)  
   ↓  
Auth Service  
   ↓  
PostgreSQL  

### Components

| Service        | Description |
|---------------|------------|
| api-gateway   | Unified entry point, JWT validation, request proxy |
| auth-service  | User registration, login, JWT generation |
| PostgreSQL    | Persistent data storage |

---

## 🔐 Authentication Flow

1. User registers via `/api/v1/users/register`
2. Password is hashed using bcrypt
3. User logs in via `/api/v1/users/login`
4. Server generates JWT (24h expiration)
5. Client includes token in request header:

Authorization: Bearer <token>

6. API Gateway validates JWT before allowing access to protected routes

---

## 🛠 Tech Stack

- **Language:** Go 1.22+
- **Framework:** Gin
- **Database:** PostgreSQL
- **ORM:** GORM
- **Authentication:** JWT (HMAC SHA256)
- **Password Security:** bcrypt
- **Architecture:** Microservices + API Gateway
- **Workspace Management:** go.work

---

## 🚀 Getting Started

### 1️⃣ Clone the repository

git clone git@github.com:yourusername/internhub.git  
cd internhub  

---

### 2️⃣ Start PostgreSQL

Make sure PostgreSQL is running locally:

CREATE DATABASE internhub;

---

### 3️⃣ Run Auth Service

cd auth-service  
go run main.go  

Runs on:  
http://localhost:8081  

---

### 4️⃣ Run API Gateway

cd api-gateway  
go run cmd/main.go  

Runs on:  
http://localhost:8080  

---

## 🧪 API Testing

### Register

curl -X POST http://localhost:8081/api/v1/users/register \
-H "Content-Type: application/json" \
-d '{"email":"test@test.com","password":"12345678","name":"test"}'

---

### Login

curl -X POST http://localhost:8081/api/v1/users/login \
-H "Content-Type: application/json" \
-d '{"email":"test@test.com","password":"12345678"}'

Returns:

{
  "access_token": "xxxxx.yyyyy.zzzzz"
}

---

### Access Protected Route

curl http://localhost:8080/api/v1/protected \
-H "Authorization: Bearer <your_token>"

---

## 📂 Project Structure

internhub/  
│  
├── api-gateway/  
│   └── cmd/main.go  
│  
├── auth-service/  
│   └── main.go  
│  
├── pkg/  
│  
├── go.work  
└── go.mod  

---

## 🎯 Key Highlights

- Designed a microservices authentication system
- Implemented stateless JWT validation at API Gateway
- Used bcrypt for secure password hashing
- Managed multi-module Go workspace
- Separated authentication logic from gateway
- Implemented middleware-based authorization

---

## 📌 Future Improvements

- Docker Compose containerization
- Role-based access control (RBAC)
- Refresh token mechanism
- Redis for token blacklist
- CI/CD pipeline integration

---

## 👨‍💻 Author

Backend engineering practice project focused on real-world architecture design.
