# Shalby HMS - Go Backend Setup Guide

## Project Overview

This is a complete authentication system built with Go for the Shalby Hospital Management System (HMS). It provides:
- User login with role-based access
- Patient registration with detailed medical information
- JWT-based token authentication
- PostgreSQL database integration
- CORS support for Vue 3 frontend

## Prerequisites

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **PostgreSQL 12+** - Running locally on port 5432
- **PostgreSQL created database** - `shalby_hospital` with 19 tables

## Quick Start

### 1. Install Dependencies

```bash
cd shalby_backend
go mod download
go mod tidy
```

### 2. Configure Environment Variables

Edit `.env` file with your database credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=shalby_hospital
DB_PASSWORD=ShalbyHMS@718
DB_NAME=shalby_hospital
DB_SSLMODE=disable
JWT_SECRET=shalby_hospital_jwt_secret_key_change_in_production_2024
JWT_EXPIRY_HOURS=24
PORT=8080
GIN_MODE=debug
```

### 3. Create Test Users (Optional)

Before running the backend, you can create test users. First, you'll need the bcrypt hashes for passwords.

#### Generate Bcrypt Hashes

Create a quick Go script to generate bcrypt hashes:

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    // Replace with passwords you want to use
    passwords := map[string]string{
        "Admin@123":       "admin@shalby.com",
        "Doctor@123":      "doctor@shalby.com",
        "Pharmacist@123":  "pharmacist@shalby.com",
        "Patient@123":     "john@example.com, jane@example.com",
    }

    for password, email := range passwords {
        hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        fmt.Printf("Password: %s (for %s)\n", password, email)
        fmt.Printf("Hash: %s\n\n", string(hash))
    }
}
```

Update `seed.sql` with the generated bcrypt hashes, then run:

```bash
psql -U shalby_hospital -d shalby_hospital -f seed.sql
```

### 4. Run the Backend

```bash
go run main.go
```

You should see:

```
✓ Configuration loaded
✓ Database connected successfully
✓ Routes setup complete
🚀 Shalby HMS Backend running on http://localhost:8080
📝 API docs available at http://localhost:8080/api/health
🔐 Authentication endpoints:
   - POST http://localhost:8080/api/auth/login
   - POST http://localhost:8080/api/auth/signup
```

## API Endpoints

### Health Check
```
GET /api/health
Response: { "status": "ok" }
```

### User Login
```
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "Patient@123",
  "role": "Patient"
}

Response (200 OK):
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "avatarUrl": "https://i.pravatar.cc/150?u=john",
    "isActive": true,
    "createdAt": "2024-03-26T10:30:00Z",
    "roles": ["Patient"]
  }
}
```

### Patient Signup
```
POST /api/auth/signup
Content-Type: application/json

{
  "fullName": "John Doe",
  "dateOfBirth": "1990-05-12",
  "gender": "Male",
  "bloodGroup": "O+",
  "phone": "+1 555-0101",
  "email": "john@example.com",
  "password": "password123",
  "allergies": ["Penicillin", "Dust"],
  "conditions": ["Diabetes"],
  "emergencyContactName": "Mary Doe",
  "emergencyContactPhone": "+1 555-0102",
  "emergencyContactRelation": "Spouse"
}

Response (201 Created):
{
  "message": "Registration successful",
  "userId": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Project Structure

```
shalby_backend/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── .env                    # Environment variables
├── seed.sql                # Test data SQL script
├── config/
│   └── config.go           # Configuration loader
├── database/
│   └── db.go               # PostgreSQL connection
├── models/
│   ├── user.go             # User and auth request/response models
│   └── patient.go          # Patient and signup request models
├── handlers/
│   └── auth.go             # Login and signup HTTP handlers
├── services/
│   └── auth.go             # Authentication business logic
├── middleware/
│   ├── cors.go             # CORS middleware
│   └── auth.go             # JWT validation middleware
└── routes/
    └── routes.go           # Route definitions
```

## Test Users (After Seed Data)

### Admin
- **Email:** admin@shalby.com
- **Password:** Admin@123
- **Role:** Admin

### Doctor
- **Email:** doctor@shalby.com
- **Password:** Doctor@123
- **Roles:** Doctor, Patient

### Pharmacist
- **Email:** pharmacist@shalby.com
- **Password:** Pharmacist@123
- **Role:** Pharmacist

### Patient 1
- **Email:** john@example.com
- **Password:** Patient@123
- **Name:** John Doe
- **DOB:** 1990-05-12
- **Blood Group:** O+
- **Allergies:** Penicillin, Dust
- **Conditions:** Diabetes
- **Emergency Contact:** Mary Doe (Spouse)

### Patient 2
- **Email:** jane@example.com
- **Password:** Patient@123
- **Name:** Jane Smith
- **DOB:** 1992-08-22
- **Blood Group:** A+
- **Allergies:** Aspirin
- **Conditions:** Hypertension
- **Emergency Contact:** Robert Smith (Brother)

## Database Schema

### Key Tables Used

- **users** - User accounts with email, password hash, and metadata
- **roles** - Available roles (Admin, Doctor, Patient, Pharmacist)
- **user_roles** - Many-to-many mapping between users and roles
- **patients** - Patient-specific information
- **patient_allergies** - Patient allergy records
- **patient_chronic_conditions** - Patient medical conditions
- **patient_emergency_contacts** - Emergency contact information

## Security Features

✅ **Password Hashing:** Bcrypt with DefaultCost
✅ **JWT Tokens:** 24-hour expiry with HMAC-SHA256
✅ **CORS Protection:** Only allows localhost:5173
✅ **SQL Injection Prevention:** Prepared statements for all queries
✅ **Password Masking:** Never returns password hashes in API responses
✅ **Transaction Safety:** Database operations are atomic

## Common Issues & Troubleshooting

### "Connection refused" error
- Ensure PostgreSQL is running
- Check DB credentials in `.env`
- Verify shalby_hospital database exists

### "Invalid JWT" error (401)
- Token has expired (24 hours)
- Token signature doesn't match JWT_SECRET
- Request missing Authorization header

### "Email already exists" error (409)
- User already registered
- Try with a different email

### "You do not have access as [Role]" (401)
- User doesn't have the requested role
- Check user_roles table for role assignment

## Frontend Integration

The Vue 3 frontend connects to this backend at:
- **API Base URL:** http://localhost:8080/api
- **Login Form:** Sends email, password, role
- **Signup Form:** Sends complete patient information
- **Token Storage:** Saved in localStorage as `hms_token`

### Frontend Setup

```bash
cd jaival_Shalby_frontend
npm install
npm run dev  # Runs on http://localhost:5173
```

## Performance Notes

- **Connection Pool:** 25 max connections, 5 idle minimum
- **Request Timeout:** 30 seconds (Gin default)
- **Token Signing:** ~10ms per token (bcrypt may take 100-200ms depending on cost)
- **Database Queries:** All use prepared statements for optimal performance

## Next Steps

After basic auth is working:

1. Add protected routes middleware
2. Implement refresh token mechanism
3. Add password reset functionality
4. Add email verification (OTP)
5. Add user profile endpoints
6. Add appointment booking endpoints
7. Add prescription endpoints

## Useful Commands

```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Run with hot reload (requires godemon)
go install github.com/cosmtrek/air@latest
air

# Test endpoints with curl
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"Patient@123","role":"Patient"}'

# Monitor database connections
psql -U shalby_hospital -d shalby_hospital \
  -c "SELECT * FROM pg_stat_activity WHERE datname = 'shalby_hospital';"
```

## Production Checklist

- [ ] Change JWT_SECRET to a strong random string
- [ ] Set GIN_MODE=release
- [ ] Enable HTTPS (add TLS certificates)
- [ ] Restrict CORS to specific domain (not localhost)
- [ ] Set up database backups
- [ ] Enable SQL query logging and monitoring
- [ ] Add rate limiting to auth endpoints
- [ ] Implement request logging and monitoring
- [ ] Set up health check alerts

## Support & Issues

For issues or questions:
1. Check the troubleshooting section above
2. Review database schema in PostgreSQL
3. Check application logs for detailed error messages
4. Verify all 19 required tables exist in shalby_hospital database

---

**Backend Created:** March 26, 2024
**Go Version:** 1.21+
**Last Updated:** 2024-03-26
