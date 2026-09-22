# Peter Go Auth

![background](./misc/images/background.png)

Production-oriented authentication API built with **Go, Gin, MongoDB, JWT, bcrypt, and Resend**.

Provides a complete authentication flow with email verification, access and refresh tokens, password recovery, role-based authorization, and rate limiting.

## Features

* **Signup** — create an account and receive an email verification code
* **Email verification** — verify accounts using a 6-digit OTP
* **Resend OTP** — request a new verification code
* **Login** — authenticate verified users and issue tokens
* **Token refresh** — obtain new access tokens using refresh tokens
* **Logout** — revoke the stored refresh token
* **Forgot password** — request a password reset email
* **Reset password** — securely set a new password
* **Protected routes** — authenticate requests using JWTs
* **Role-based access** — restrict endpoints by user role
* **Admin routes** — dedicated admin-only authorization
* **Rate limiting** — IP-based protection for public authentication endpoints
* **User enumeration protection** — consistent responses for sensitive email-based flows

## Tech Stack

* **Go** + **Gin** — HTTP API
* **MongoDB** — user persistence
* **JWT (HMAC-SHA256)** — access, refresh, and reset tokens
* **bcrypt** — password hashing
* **Resend** — transactional email
* **`golang.org/x/time/rate`** — rate limiting

## Requirements

Before getting started, make sure you have:

* Go 1.22+
* MongoDB
* A [Resend](https://resend.com/) account and API key

## Installation

Clone the repository:

```bash
git clone https://github.com/GlitchGuru19/peter-go-auth-template.git
cd peter-go-auth-template
```

Install dependencies:

```bash
go mod tidy
```

## Configuration

Create a `.env` file in the project root:

```env
PORT=8000

MONGODB_URI=mongodb://localhost:27017
SECRET_KEY=your_long_random_secret

RESEND_API_KEY=re_your_key
RESEND_FROM_EMAIL="Peter Auth <onboarding@resend.dev>"

FRONTEND_URL=http://localhost:3000
```

### Environment Variables

| Variable            | Description                                            |
| ------------------- | ------------------------------------------------------ |
| `PORT`              | Port the API listens on                                |
| `MONGODB_URI`       | MongoDB connection string                              |
| `SECRET_KEY`        | Secret used to sign JWTs                               |
| `RESEND_API_KEY`    | Resend API key                                         |
| `RESEND_FROM_EMAIL` | Email address used to send transactional emails        |
| `FRONTEND_URL`      | Frontend URL used when generating password-reset links |

> **Security:** Never commit `.env` files, API keys, JWT secrets, or other credentials to version control.

## Database Setup

Create a unique index on the `email` field in the `users` collection:

```javascript
db.users.createIndex(
  { email: 1 },
  { unique: true }
)
```

This ensures that each email address can only belong to one account.

No TTL index should be created for password-reset fields because those fields are stored directly on the user document. A TTL index on those fields would delete the entire user document when the expiration date is reached.

## Running the API

Start the server:

```bash
go run main.go
```

The API will be available at:

```text
http://localhost:8000
```

## Authentication

### Signup & Email Verification

The registration flow is:

```text
Signup
   ↓
6-digit OTP
   ↓
Email verification
   ↓
Login
   ↓
Access + Refresh tokens
```

OTP codes are:

* Generated using `crypto/rand`
* Stored as SHA-256 hashes
* Valid for 10 minutes
* Limited to 5 verification attempts

Unverified accounts cannot log in.

### Access & Refresh Tokens

Successful authentication returns:

* **Access token** — used for authenticated API requests
* **Refresh token** — used to obtain a new access token

JWTs contain a `type` claim:

```text
access
refresh
reset
```

Token types are validated independently so that a token cannot be used outside its intended purpose.

### Password Reset

The password recovery flow is:

```text
Forgot password
      ↓
Reset email
      ↓
Reset link
      ↓
Token validation
      ↓
New password
```

Password-reset tokens:

* Expire after 15 minutes
* Are single-use
* Are cleared after successful password reset

## API

| Method | Endpoint           | Description              | Access        |
| ------ | ------------------ | ------------------------ | ------------- |
| `POST` | `/signup`          | Create an account        | Public        |
| `POST` | `/verify-otp`      | Verify email address     | Public        |
| `POST` | `/send-otp`        | Resend verification code | Public        |
| `POST` | `/login`           | Authenticate a user      | Public        |
| `POST` | `/refresh`         | Refresh an access token  | Public        |
| `POST` | `/logout`          | Revoke refresh token     | Authenticated |
| `POST` | `/forgot-password` | Request password reset   | Public        |
| `POST` | `/reset-password`  | Set a new password       | Public        |
| `GET`  | `/me`              | Get authenticated user   | Authenticated |
| `GET`  | `/admin`           | Admin-only endpoint      | Admin         |

## Security

### Password Security

Passwords are hashed with bcrypt before being stored.

Plaintext passwords are never persisted.

### JWT Security

JWTs are signed using HMAC-SHA256 and include an explicit token type.

Access, refresh, and password-reset tokens are validated separately.

### OTP Security

OTP codes are generated using a cryptographically secure random source and stored only as SHA-256 hashes.

Each code expires after 10 minutes and is limited to 5 attempts.

### Password Reset Security

Password-reset tokens are short-lived and single-use.

After a successful password reset, the reset token is cleared.

### User Enumeration Protection

Sensitive email-based endpoints return consistent responses regardless of whether an email address exists.

This applies to:

* `/forgot-password`
* `/send-otp`

### Rate Limiting

Public authentication endpoints are protected with IP-based rate limiting to reduce brute-force attempts and automated abuse.

## Production Deployment

Before deploying to production, review the following.

### Refresh Tokens

Refresh tokens are currently stored in the database in raw form for simplicity.

For a higher-security deployment, consider implementing:

* Hashed refresh-token storage
* Refresh-token rotation
* Token reuse detection
* Session/device management
* Explicit token expiration and revocation

### Rate Limiting

The current rate limiter is in-memory and works per application instance.

If running multiple instances, use a shared store such as Redis so rate limits are consistent across the deployment.

### Email

For development, Resend allows:

```env
RESEND_FROM_EMAIL="Peter Auth <onboarding@resend.dev>"
```

For production, verify your own domain with Resend and use an address from that domain.

### Secrets

Production secrets should be provided through environment variables or a dedicated secrets manager.

Never commit production credentials to the repository.

## License

This project is licensed under the [MIT License](LICENSE).
