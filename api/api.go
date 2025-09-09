package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
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

// HandleCreateShortenURL godoc
//
// @Summary      Create a shortened URL
// @Description  Takes an original URL, validates it, and stores it in the database.
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        request body createShortenURLRequest true "Original URL request"
// @Success      201 {object} createShortenURLResponse "Shortened URL created successfully"
// @Failure      400 {object} map[string]string "Invalid input or URL already exists"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /api/urls [post]
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
	shortenURL := service.GenerateShortenURL(server.config, res.ID)
	server.logger.Info("Create URL shorten successfully", "url", shortenURL)
	resp := createShortenURLResponse{
		ShortenURL: shortenURL,
	}

	// Write response back to client
	server.WriteJSON(w, http.StatusCreated, resp)
}

// HandleRedirect godoc
// @Summary      Redirect to original URL
// @Description  Redirects a visitor from the shortened URL code to the original URL and records the visit.
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        code path string true "Shortened URL code"
// @Success      301 {string} string "Redirected successfully"
// @Failure      400 {object} map[string]string "Invalid code or URL not found"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /{code} [get]
func (server *Server) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	// Get and decode the ID
	id := service.DecodeBase62(r.PathValue("code"))

	// Get the original URL in the database
	url, err := server.queries.GetURL(r.Context(), id)
	if err != nil {
		// If the ID is invalid (not match any record)
		if errors.Is(err, sql.ErrNoRows) {
			server.WriteError(w, http.StatusBadRequest, "This URL didn't existed")
			return
		}

		// Other database errors
		server.logger.Error("GET /{code}: failed to get original URL", "error", err)
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
		server.logger.Error("GET /{code}: failed to record the visitor", "error", err)
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

// HandleListURL godoc
// @Summary      List registered URLs
// @Description  Retrieves a paginated list of all shortened URLs in the system.
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        page_size  query int true  "Number of items per page" minimum(1) maximum(100)
// @Param        page_index query int true  "Page index (starting from 1)" minimum(1)
// @Success      200 {array} listURLResponse "List of shortened URLs"
// @Failure      400 {object} map[string]string "Invalid pagination parameters"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /api/urls [get]
func (server *Server) HandleListURL(w http.ResponseWriter, r *http.Request) {
	// Get the page_size and page_index parameter
	pageSize, pageIndex, err := server.ExtractPageParams(r)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the list
	urls, err := server.queries.ListURL(r.Context(), db.ListURLParams{
		Offset: (pageIndex - 1) * pageSize,
		Limit:  pageSize,
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

// HandleListVisitor godoc
// @Summary      List visitors for a shortened URL
// @Description  Retrieves a paginated list of visitors who accessed the given shortened URL.
// @Tags         visitors
// @Accept       json
// @Produce      json
// @Param        id          path  string true  "Shortened URL ID (base62 code)"
// @Param        page_size   query int    true  "Number of items per page" minimum(1) maximum(100)
// @Param        page_index  query int    true  "Page index (starting from 1)" minimum(1)
// @Success      200 {array} listVisitorResponse "List of visitors"
// @Failure      400 {object} map[string]string "Invalid pagination parameters"
// @Failure      404 {object} map[string]string "URL ID not found"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /api/urls/{id}/visitors [get]
func (server *Server) HandleListVisitor(w http.ResponseWriter, r *http.Request) {
	// Get URL ID from path parameter
	id := service.DecodeBase62(r.PathValue("id"))

	// Get the page_size and page_index parameter
	pageSize, pageIndex, err := server.ExtractPageParams(r)
	if err != nil {
		server.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the list of visitor who had visit to this URL
	visitors, err := server.queries.ListVisitor(r.Context(), db.ListVisitorParams{
		UrlID:  id,
		Offset: (pageIndex - 1) * pageSize,
		Limit:  pageSize,
	})

	if err != nil {
		// If ID not match any record
		if errors.Is(err, sql.ErrNoRows) {
			server.WriteError(w, http.StatusNotFound, "This URL ID does not match any record")
			return
		}

		// If other database errors
		server.logger.Error("GET /api/urls/{id}/visitors: failed to get the list of visitor for this url",
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

// HandleCountURL godoc
// @Summary      Get total URLs
// @Description  Returns the total number of shortened URLs stored in the system (useful for pagination).
// @Tags         urls
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]int64 "Total number of URLs"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /api/urls/count [get]
func (server *Server) HandleCountURL(w http.ResponseWriter, r *http.Request) {
	count, err := server.queries.CountURL(r.Context())
	if err != nil {
		server.logger.Error("GET /api/urls/count: failed to get the total of the URLs in database")
		server.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	server.WriteJSON(w, http.StatusOK, map[string]int64{
		"total_urls": count,
	})
}
