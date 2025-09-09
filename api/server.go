package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	db "github.com/danglnh07/URLShortener/db/sqlc"
	_ "github.com/danglnh07/URLShortener/docs"
	"github.com/danglnh07/URLShortener/service"
	"github.com/go-playground/validator/v10"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server struct, holds all dependency used for backend, config and logger
type Server struct {
	mux      *http.ServeMux
	config   *service.Config
	queries  *db.Queries
	validate *validator.Validate
	limiter  *RateLimiter
	logger   *slog.Logger
}

// Constructor method for Server
func NewServer(config *service.Config, conn *sql.DB, logger *slog.Logger) *Server {
	return &Server{
		mux:      http.NewServeMux(),
		config:   config,
		queries:  db.New(conn),
		validate: validator.New(validator.WithRequiredStructEnabled()),
		limiter:  NewRateLimiter(config.MaxRequest, config.RefillRate),
		logger:   logger,
	}
}

// Helper method for registering handler
func (server *Server) RegisterHandler() {
	// Register API handlers
	server.mux.Handle("GET /api/urls/{id}/visitors", http.Handler(
		server.ChainingMiddleware(http.HandlerFunc(server.HandleListVisitor))),
	)
	server.mux.Handle("POST /api/urls", http.Handler(
		server.ChainingMiddleware(http.HandlerFunc(server.HandleCreateShortenURL))),
	)
	server.mux.Handle("GET /api/urls/count", http.Handler(
		server.ChainingMiddleware(http.HandlerFunc(server.HandleCountURL))),
	)
	server.mux.Handle("GET /api/urls", http.Handler(
		server.ChainingMiddleware(http.HandlerFunc(server.HandleListURL))),
	)

	// Shorten URL handling
	server.mux.Handle("GET /{code}", http.Handler(
		server.ChainingMiddleware(http.HandlerFunc(server.HandleRedirect))),
	)

	// Swagger handler
	server.mux.Handle("/swagger/", httpSwagger.WrapHandler)
}

// Method to start the server
func (server *Server) Start() error {
	// Register handler
	server.RegisterHandler()

	// Startserver
	server.logger.Info("Starting server", "address",
		fmt.Sprintf("%s:%s", server.config.Domain, server.config.Port))
	return http.ListenAndServe(fmt.Sprintf(":%s", server.config.Port), server.mux)
}

// WriteError writes an error response in JSON format
func (server *Server) WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// WriteJSON writes a JSON response with the given status code and data in any data type
func (server *Server) WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper method to extract the pagination parameters. page_index are 1-based index
func (server *Server) ExtractPageParams(r *http.Request) (int32, int32, error) {
	// Get the page_size and page_index parameter
	params := r.URL.Query()
	pageSize, err := strconv.Atoi(params.Get("page_size"))
	if err != nil {
		return -1, -1, fmt.Errorf("invalid value for page_size: %v", err)
	}
	if pageSize <= 0 || pageSize > 100 {
		return -1, -1, fmt.Errorf(
			"invalid value for page_size, must be a positive integer smaller than or equal 100",
		)
	}

	pageIndex, err := strconv.Atoi(params.Get("page_index"))
	if err != nil {
		return -1, -1, fmt.Errorf("invalid value for page_index")
	}
	if pageIndex <= 0 {
		return -1, -1, fmt.Errorf("invalid value for page_size, must be a positive integer")
	}

	return int32(pageSize), int32(pageIndex), nil
}
