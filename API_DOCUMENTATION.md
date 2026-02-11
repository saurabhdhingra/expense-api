# Expense Tracker API Documentation

The Expense Tracker API is a RESTful service built with Go and the Gin framework. It provides endpoints for user authentication, wallet management, expense logging, and financial insights.

## Base URL
The API runs locally on `http://localhost:8080`.

## Authentication
Most endpoints require a JSON Web Token (JWT) for authentication.
- **Header**: `Authorization`
- **Format**: `Bearer <your_jwt_token>`

Tokens are obtained via the `/register` or `/login` endpoints and are valid for 24 hours.

---

## Public Endpoints

### 1. Health Check
Check the status of the API.

- **URL**: `/health`
- **Method**: `GET`
- **Auth required**: No
- **Success Response**: `{"status": "UP", "message": "Expense Tracker API is running"}`

### 2. Register User
Create a new user account. Returns a JWT token.

- **URL**: `/register`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }
  ```

### 3. Login User
Authenticate an existing user. Returns a JWT token.

- **URL**: `/login`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "email": "john@example.com",
    "password": "securepassword123"
  }
  ```

---

## Wallet Endpoints (Authenticated)

### 4. Create Wallet
Create a new wallet (Cash, Debit Card, etc.).

- **URL**: `/wallets`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "name": "Primary Bank",
    "currency": "USD",
    "balance": 5000.0,
    "type": "Debit Card",
    "company": "Visa",
    "color": "blue"
  }
  ```

### 5. List Wallets
Retrieve all wallets for the authenticated user.

- **URL**: `/wallets`
- **Method**: `GET`
- **Success Response**: Array of wallet objects with string IDs.

---

## Expense Endpoints (Authenticated)

### 6. Create Expense
Log a new expense and automatically deduct from wallet balance if `wallet_id` is provided.

- **URL**: `/expenses`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "description": "Starbucks Coffee",
    "amount": 5.50,
    "category": "Leisure",
    "date": "2025-02-11T10:00:00Z",
    "wallet_id": 1
  }
  ```

### 7. List Expenses (Grouped for iOS)
Retrieve expenses. When `filter=all` (default), returns grouped data for the home screen.

- **URL**: `/expenses`
- **Method**: `GET`
- **Success Response (Default)**:
  ```json
  {
    "monthlyExpenses": [
      {
        "month": "February 2025",
        "expenses": [...]
      }
    ],
    "currentMonthDistribution": [
      {
        "category": "Groceries",
        "totalAmount": 150.0
      }
    ]
  }
  ```

---

## Insights Endpoints (Authenticated)

### 8. Get Insights
Retrieve chart data and top transactions for a specific period.

- **URL**: `/analytics`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "interval": "monthly",
    "walletId": "",
    "category": ""
  }
  ```
  *Intervals: `daily` (last 7 days), `monthly` (last 12 months), `yearly` (last 5 years).*
- **Success Response**:
  ```json
  {
    "chartData": [
      { "id": "Jan", "label": "Jan", "amount": 1200.0 }
    ],
    "topTransactions": [...]
  }
  ```
