package backend

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "github.com/danglnh07/URLShortener/db/sqlc"
	"github.com/danglnh07/URLShortener/service"
)

// request struct for create shorten URL action
type createShortenURLRequest struct {
	URL string `json:"url" validate:"required"`
}

// response struct for create shorten URL action
type createShortenURLResponse struct {
	ShortenURL string `json:"shorten_url"`
}

// Handler for create shorten URL action
// endpoint: POST /api/urls
// Success: 201
// Fail: 400, 500
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
		server.logger.Error("POST /api/urls: failed to insert URL into database", "error", err)
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

// Handler for redirect from shorten URL to original URL
// endpoint: GET /
// Success: 301
// Fail: 400, 500
func (server *Server) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	// Get and decode the ID
	id := service.DecodeBase62(r.URL.Path)

	// Get the original URL in the database
	url, err := server.queries.GetURL(r.Context(), id)
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
		UrlID: id,
	})
	if err != nil {
		server.logger.Error("GET /: failed to record the visitor", "error", err)
		// Should NOT return an error here
	}

	// Redirect to the original URL
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

// Response struct for listing URLs
type listURLResponse struct {
	OriginalURL  string    `json:"original"`
	ShortenURL   string    `json:"shorten"`
	TotalVisitor int64     `json:"total_visitor"`
	CreatedAt    time.Time `json:"created_at"`
}

// Handler for listing all URLs that has been registered in the system
// endpoint: GET /api/urls?page_size=...&page_index=...
// Success: 200
// Fail: 400, 500
func (server *Server) HandleListURL(w http.ResponseWriter, r *http.Request) {
	// Get the page_size and page_index parameter
	pageSize, pageIndex, err := server.ExtractPageParams(r)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the list
	urls, err := server.queries.ListURL(r.Context(), db.ListURLParams{
		Offset: int32((pageIndex - 1) * pageSize),
		Limit:  int32(pageSize),
	})
	if err != nil {
		server.logger.Error("GET /api/urls?page_size=...&page_index=...: failed to get list of URLS",
			"error", err)
		server.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create response struct
	resps := make([]listURLResponse, len(urls))
	for i, url := range urls {
		resps[i] = listURLResponse{
			OriginalURL:  url.OriginalUrl,
			ShortenURL:   service.GenerateShortenURL(server.config, url.ID),
			TotalVisitor: url.TotalVisitors,
			CreatedAt:    url.TimeCreated,
		}
	}

	// Return the list
	server.WriteJSON(w, http.StatusOK, resps)

}

// Response struct for list visitor for each URL action
type listVisitorResponse struct {
	Ip          string    `json:"ip"`
	OriginalURL string    `json:"original"`
	ShortenURL  string    `json:"shorten"`
	TimeVisited time.Time `json:"time_visited"`
}

// Handler for listing all visitor that has visit the URL
// endpoint: GET /api/urls/{id}/visitors?page_size=...&page_index=...
// Success: 200
// Fail: 400, 404, 500
func (server *Server) HandleListVisitor(w http.ResponseWriter, r *http.Request) {
	// Get URL ID from path parameter
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, "Invalid URL ID")
		return
	}

	// Get the page_size and page_index parameter
	pageSize, pageIndex, err := server.ExtractPageParams(r)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the list of visitor who had visit to this URL
	visitors, err := server.queries.ListVisitor(r.Context(), db.ListVisitorParams{
		UrlID:  id,
		Offset: int32((pageIndex - 1) * pageSize),
		Limit:  int32(pageSize),
	})

	if err != nil {
		// If ID not match any record
		if errors.Is(err, sql.ErrNoRows) {
			server.WriteError(w, http.StatusNotFound, "This URL ID does not match any record")
			return
		}

		// If other database errors
		server.logger.Error("GET /urls/{id}: failed to get the list of visitor for this url",
			"url_id", id, "error", err)
		server.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create response list
	resps := make([]listVisitorResponse, len(visitors))
	for i, visitor := range visitors {
		resps[i] = listVisitorResponse{
			Ip:          visitor.Ip,
			OriginalURL: visitor.OriginalUrl,
			ShortenURL:  service.GenerateShortenURL(server.config, visitor.UrlID),
			TimeVisited: visitor.TimeVisited,
		}
	}

	// Return result to client
	server.WriteJSON(w, http.StatusOK, resps)
}
