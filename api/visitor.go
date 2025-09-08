package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	db "github.com/danglnh07/URLShortener/db/sqlc"
)

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

	// Return result to client
	server.WriteJSON(w, http.StatusOK, visitors)
}
