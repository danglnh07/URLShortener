package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	db "github.com/danglnh07/URLShortener/db/sqlc"
	"github.com/danglnh07/URLShortener/service"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	mux      *http.ServeMux
	config   *service.Config
	queries  *db.Queries
	validate *validator.Validate
	logger   *slog.Logger
}

func NewServer(config *service.Config, conn *sql.DB, logger *slog.Logger) *Server {
	return &Server{
		mux:      http.NewServeMux(),
		config:   config,
		queries:  db.New(conn),
		validate: validator.New(validator.WithRequiredStructEnabled()),
		logger:   logger,
	}
}

func (server *Server) RegisterHandler() {
	server.mux.HandleFunc("POST /urls", server.HandleCreateShortenURL)
	server.mux.HandleFunc("GET /", server.HandleRedirect)
}

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
		"message": message,
	})
}

// WriteJSON writes a JSON response with the given status code and data in any data type
func (server *Server) WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"data": data,
	})
}
