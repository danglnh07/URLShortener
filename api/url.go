package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	db "github.com/danglnh07/URLShortener/db/sqlc"
	"github.com/danglnh07/URLShortener/service"
)

type createShortenURLRequest struct {
	URL string `json:"url" validate:"required"`
}

type createShortenURLResponse struct {
	ShortenURL string `json:"shorten_url"`
}

func (server *Server) HandleCreateShortenURL(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request and validate
	var req createShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := server.validate.Struct(req); err != nil {
		server.WriteError(w, http.StatusBadRequest, "url should not be empty")
		return
	}

	// Insert URL into database
	res, err := server.queries.CreateURL(r.Context(), req.URL)
	if err != nil {
		// If URL already exists in database
		if strings.Contains(err.Error(), "url_original_url_key") {
			server.WriteError(w, http.StatusBadRequest, "This URL has been registered")
			return
		}

		// Other database errors
		server.logger.Error("POST /urls: failed to insert URL into database", "error", err)
		server.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create response with shorten URL using the database ID
	shortenURL := fmt.Sprintf("%s:%s/%s",
		server.config.Domain, server.config.Port, service.EncodeBase62(int64(res.ID)))
	server.logger.Info("Create URL shorten successfully", "url", shortenURL)
	resp := createShortenURLResponse{
		ShortenURL: shortenURL,
	}

	// Write response back to client
	server.WriteJSON(w, http.StatusCreated, resp)
}

func (server *Server) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	// Get and decode the ID
	id := service.DecodeBase62(r.URL.Path)

	// Get the original URL in the database
	url, err := server.queries.GetURL(r.Context(), int32(id))
	if err != nil {
		// If the ID is invalid (not match any record)
		if errors.Is(err, sql.ErrNoRows) {
			server.WriteError(w, http.StatusBadRequest, "This URL didn't existed")
			return
		}

		// Other database errors
		server.logger.Error("GET /: failed to get original URL", "error", err)
		server.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Get the visitor IP address
	ip := ""
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = strings.Split(fwd, ",")[0]
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	server.logger.Info("Visitor info", "IP", ip)

	// Record the visitor
	_, err = server.queries.CreateVisitor(r.Context(), db.CreateVisitorParams{
		Ip:    ip,
		UrlID: int32(id),
	})
	if err != nil {
		server.logger.Error("GET /: failed to record the visitor")
		// Should NOT return an error here
	}

	// Redirect to the original URL
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
