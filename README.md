💰 Go RESTful Expense Tracker API

This project provides a secure backend API for a personal expense tracking application. It is built with Go (Golang), the Gin framework, and uses PostgreSQL with the GORM ORM.

✨ Features

User Authentication: Secure registration and login using JWT (JSON Web Tokens).

Secure Storage: Passwords hashed with bcrypt.

Expense CRUD: Full management of expense records (Create, Read, Update, Delete).

Ownership: Strict authorization ensures users only interact with their own expenses.

Filtering: List expenses by predefined (week, month, 3 months) or custom date ranges.

Analytics: Dedicated endpoint to map daily, monthly, and yearly expenses, providing total spend and the top 5 largest transactions for each period.

Expense Categories: Expenses are validated against a predefined list of categories (Groceries, Leisure, Utilities, etc.).

🛠️ Prerequisites

Go: Version 1.21 or higher.

PostgreSQL: A running instance of a PostgreSQL server.

Git: For cloning the repository.

🚀 Getting Started

1. Setup

# Clone the repository
git clone <repository-url>
cd expense-tracker-api

# Install dependencies
go mod tidy


2. Database Configuration

You must update the Database Source Name (DSN) in config/database.go to connect to your local PostgreSQL server. GORM will automatically create the users and expenses tables upon startup.

// Inside config/database.go
dsn := "host=localhost user=user password=password dbname=expensetracker port=5432 sslmode=disable TimeZone=Asia/Shanghai" 


3. Run the Application

go run main.go


The server will start on port 8080.

🔌 API Endpoints

The base URL is http://localhost:8080.

Authentication (Public)

Endpoint

Method

Body Parameters

Description

/register

POST

name, email, password

Create a new user account. Returns a JWT token.

/login

POST

email, password

Authenticate and return a JWT token.

Expense Management (Requires Authorization: Bearer <TOKEN>)

Endpoint

Method

Parameters

Description

/expenses

POST

description, amount, category, date

Create a new expense.

/expenses/:id

PUT

Partial expense body

Update an existing expense by ID.

/expenses/:id

DELETE

(None)

Delete an expense by ID.

/analytics

GET

(None)

Returns daily, monthly, and yearly expense summaries (total and top 5 transactions for each period).

Listing and Filtering Examples (GET /expenses)

The GET /expenses endpoint supports filtering:

Filter Type

Query Parameters

Example

Past Week

?filter=past_week

/expenses?filter=past_week

Past Month

?filter=past_month

/expenses?filter=past_month

Last 3 Months

?filter=last_3_months

/expenses?filter=last_3_months

Custom Range

?filter=custom&start_date=YYYY-MM-DD&end_date=YYYY-MM-DD

/expenses?filter=custom&start_date=2023-10-01&end_date=2023-10-31

All (Default)

(None)

/expenses