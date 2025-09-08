URL shortener built using Go

# Features

- Convert long URL to short URL
- Redirect shorten URL to original URL
- Track the total number of visitor to the URL
- Track IP addresses of visitor who vist the URL
- Simple web UI

# Tech stack

- Database: Postgres 17, `sqlc` for query generation
- `net/http` package for building API
- `Makefile` for build tool
- `Docker` for containerization

# How to use this project

1. Navigate to your working directory
2. Clone this project

```bash
git clone https://github.com/danglnh07/URLShortener.git
```
3. Run the project

```bash
# If using make
make run

# If using Go tool
go run main.go
```
