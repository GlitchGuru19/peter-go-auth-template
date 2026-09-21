# Implementation Guide

## Database Connection Setup

### Step 1: Install Required Packages
```bash
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/mongo/options
```

### Step 2: Add Environment Variables to .env
Add these variables to your `.env` file:
```
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=your_database_name
```

### Step 3: Load Database Environment Variables
In `initializers/loadEnvVariables.go`, update to load the new MongoDB variables alongside the existing ones.

### Step 4: Create Database Initializer File
Create `initializers/database.go` with the following structure:
- Import `go.mongodb.org/mongo-driver/mongo`, `go.mongodb.org/mongo-driver/mongo/options`, `context`, and `fmt`
- Create a var to store the DB connection: `var DB *mongo.Client`
- Create a function `ConnectToDb()` that:
  - Gets MONGODB_URI from environment variables
  - Creates a context with timeout
  - Creates MongoDB client options with the URI
  - Connects using `mongo.Connect(ctx, clientOpts)`
  - Pings the database to verify connection
  - Handles errors
  - Assigns to the global `DB` variable
  - Note: Don't forget to defer cancel() after creating context

### Step 5: Call ConnectToDb in main.go
- Call `ConnectToDb()` after `LoadEnvVariables()` in your `init()` function
- This ensures database connection happens during initialization

### Step 6: Test the Connection
Run `go run main.go` to verify the database connects without errors.

### MongoDB Connection URI Format Reference
```
mongodb://localhost:27017
mongodb+srv://username:password@cluster.mongodb.net/?retryWrites=true&w=majority
```

### Common Error Handling
- Connection refused: Check if MongoDB is running on localhost:27017
- Authentication failed: Verify credentials in URI match your MongoDB setup
- Database connection timeout: Increase timeout duration in context
- Ping failed: Ensure MongoDB service is accessible

## Test Signup and Login Endpoints

Run the app first:
```bash
go run main.go
```

By default, the API runs on `http://localhost:8000` unless you set a different `PORT` in your `.env` file.

#### 1) Sign up a new user
Use a `POST` request to `/signup` with JSON like this:

```bash
curl -X POST http://localhost:8000/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "Password123!"
  }'
```

Expected response:
```json
{
  "message": "Account created successfully",
  "user_id": "...mongodb-object-id..."
}
```

Status code: `201 Created`

#### 2) Log in with the same user
Use a `POST` request to `/login`:

```bash
curl -X POST http://localhost:8000/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "Password123!"
  }'
```

Expected response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Status code: `200 OK`

The returned token is the JWT that should be used in later protected routes as a bearer token.

#### 3) Test in Postman
- Open Postman
- Create a new `POST` request
- Set URL: `http://localhost:8000/signup`
- Go to `Body` -> `raw` -> `JSON`
- Paste:
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "Password123!"
}
```
- Click `Send`
- Repeat for `/login` with:
```json
{
  "email": "john@example.com",
  "password": "Password123!"
}
```

#### 4) Example `.env` values for testing
```env
PORT=8000
MONGODB_URI=mongodb://localhost:27017
SECRET_KEY=your_super_secret_key_here
```

#### 5) Common test cases
- Signup with a duplicate email should return `409 Conflict`
- Login with a wrong password should return `401 Unauthorized`
- Invalid JSON or missing required fields should return `400 Bad Request`
- Use a valid JWT token in the `Authorization` header for protected routes later

```json
Authorization: Bearer <your_token_here>
```

This is the basic flow for testing the authentication system end-to-end.

## Test the `/me` Endpoint (With and Without Token)

The `/me` endpoint typically returns the current user's profile and requires a valid JWT in the `Authorization` header.

Run the app first:
```bash
go run main.go
```

1) Call `/me` with a valid token (curl):

```bash
# replace <token> with the JWT returned from /login
curl -X GET http://localhost:8000/me \
  -H "Authorization: Bearer <token>"
```

Expected success response (HTTP `200 OK`):
```json
{
  "id": "...",
  "name": "John Doe",
  "email": "john@example.com",
  "role": "USER"
}
```

2) Call `/me` without a token (curl):

```bash
curl -X GET http://localhost:8000/me
```

Expected failure response: HTTP `401 Unauthorized` (or `403 Forbidden` depending on middleware).
Example body:
```json
{ "error": "Missing or invalid token" }
```

3) Call `/me` with an invalid/expired token (curl):

```bash
curl -X GET http://localhost:8000/me \
  -H "Authorization: Bearer invalid.or.expired.token"
```

Expected failure: HTTP `401 Unauthorized` with an error about invalid/expired token.

4) Test in Postman:
- Create a new `GET` request to `http://localhost:8000/me`.
- Method A (recommended): open the `Authorization` tab, set `Type` to `Bearer Token`, paste the token into the `Token` field, then click `Send`.
- Method B: open the `Headers` tab and add `Authorization` = `Bearer <token>`.
- To test the "without token" case, remove the Authorization header or switch Authorization type to `No Auth` and click `Send`.

5) Quick automation tip (bash): store the token and reuse it:

```bash
TOKEN=$(curl -s -X POST http://localhost:8000/login -H "Content-Type: application/json" -d '{"email":"john@example.com","password":"Password123!"}' | jq -r .token)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/me
```

Common checks:
- Missing header -> `401`
- Invalid/expired token -> `401`
- Valid token -> `200` and user profile JSON

