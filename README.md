# URL shortener built using Go

## Features

- Convert long URL to short URL
- Redirect shorten URL to original URL
- Track the total number of visit to the URL
- Track IP addresses of visitor who vist the URL

## Tech stack

- `Go v1.24.6` as the main programming language, `net/http` standard library for building API, `testing` and `httptest` as API testing tool
- `Postgres 17.5` as database, `sqlc` for database queries generation
- `Makefile` for build tool
- `Docker` for containerization
- `Swagger - swaggo` as API documentation

## How to run this project

Navigate to your working directory, then

```bash
# Clone the project
git clone https://github.com/danglnh07/URLShortener.git

# Run the container
docker compose up
```

You can also config how the app run by create an `.env` file with these value:

```bash
MAX_REQUEST=100
REFILL_RATE=10 # Second
PORT=9090 
```