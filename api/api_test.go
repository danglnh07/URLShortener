package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"

	db "github.com/danglnh07/URLShortener/db/sqlc"
	"github.com/danglnh07/URLShortener/service"
	"github.com/stretchr/testify/require"
)

var (
	server *Server
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	config service.Config
)

func TestMain(m *testing.M) {
	// Load config
	err := service.LoadConfig("../.env")
	if err != nil {
		logger.Error("Failed to load config for API testing", "error", err)

		// In CI/CD, we can get the enviroment from other source, so we don't return here
	}

	config = service.Config{
		Domain:   os.Getenv("DOMAIN"),
		Port:     os.Getenv("PORT"),
		DbDriver: os.Getenv("DB_DRIVER"),
		DbSource: os.Getenv("DB_SOURCE"),
	}

	// Connect to database
	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Create server
	server = NewServer(&config, conn, logger)
	server.RegisterHandler()

	os.Exit(m.Run())
}

func TestHandleCreateShortenURL(t *testing.T) {
	data := []string{
		"https://www.youtube.com/watch?v=LCfEqudu4pc&list=RDLCfEqudu4pc&start_radio=1&ab_channel=ForestOfLight",
		"https://www.youtube.com/watch?v=WtXzFNogegY&ab_channel=TNE",
		"https://www.youtube.com/watch?v=b2ZiE_8tPdg&ab_channel=FloWoelki",
	}

	// Run each test case
	for _, url := range data {
		// Create request
		var buffer bytes.Buffer
		err := json.NewEncoder(&buffer).Encode(createShortenURLRequest{URL: url})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/urls", &buffer)
		require.NoError(t, err)

		// Mock HTTP and test
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.HandleCreateShortenURL)
		handler.ServeHTTP(rr, req)

		// Compare result
		require.Equal(t, rr.Code, 201)
		require.NotEmpty(t, rr.Body)
	}

	// Clean up database
	err := server.queries.DeleteURL(context.Background(), data[0])
	require.NoError(t, err)
	err = server.queries.DeleteURL(context.Background(), data[1])
	require.NoError(t, err)
	err = server.queries.DeleteURL(context.Background(), data[2])
	require.NoError(t, err)
}

func TestHandleCountURL(t *testing.T) {
	// Create request
	req, err := http.NewRequest("GET", "/api/urls/count", nil)
	require.NoError(t, err)

	// Mock HTTP and test
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleCountURL)
	handler.ServeHTTP(rr, req)

	// Compare result
	require.Equal(t, rr.Code, 200)
	require.NotEmpty(t, rr.Body)

	var resp map[string]int64
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.GreaterOrEqual(t, int(resp["total_urls"]), 0)
}

func TestHandleListURL(t *testing.T) {
	// Create request
	req, err := http.NewRequest("GET", "/api/urls?page_size=5&page_index=1", nil)
	require.NoError(t, err)

	// Mock HTTP and test
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleListURL)
	handler.ServeHTTP(rr, req)

	// Compare result
	require.Equal(t, rr.Code, 200)
	require.NotEmpty(t, rr.Body)

	var resp []listURLResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.LessOrEqual(t, len(resp), 5)
}

func TestHandleListVisitor(t *testing.T) {
	// Create a shorten URL
	data := "https://www.youtube.com/watch?v=GCnipL4T0Ho&ab_channel=CodingwithSphere"

	// Create request
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(createShortenURLRequest{URL: data})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/urls", &buffer)
	require.NoError(t, err)

	// Mock HTTP and test
	createRecoder := httptest.NewRecorder()
	createHandler := http.HandlerFunc(server.HandleCreateShortenURL)
	createHandler.ServeHTTP(createRecoder, req)

	// Compare result
	require.Equal(t, createRecoder.Code, 201)
	require.NotEmpty(t, createRecoder.Body)

	// Make request to the shorten URL (in this case, the IP should be 127.0.0.x since this is localhost)
	var shortenURL createShortenURLResponse
	err = json.NewDecoder(createRecoder.Body).Decode(&shortenURL)
	require.NoError(t, err)

	u, err := url.Parse(shortenURL.ShortenURL)
	require.NoError(t, err)

	// extract code (e.g. "/abc123" -> "abc123")
	code := strings.TrimPrefix(u.Path, "/")
	req = httptest.NewRequest(http.MethodGet, "http://localhost:8080/"+code, nil)
	req.RemoteAddr = "127.0.0.1:12345"

	redirectRecoder := httptest.NewRecorder()
	server.mux.ServeHTTP(redirectRecoder, req)
	require.Equal(t, 301, redirectRecoder.Code)

	// Get the list of visitor
	req, err = http.NewRequest("GET", fmt.Sprintf("/api/urls/%s/visitors?page_size=5&page_index=1", code), nil)
	require.NoError(t, err)

	// Mock HTTP and test
	listRecorder := httptest.NewRecorder()
	server.mux.ServeHTTP(listRecorder, req)

	require.Equal(t, listRecorder.Code, 200)
	require.NotEmpty(t, listRecorder.Body)

	var resp []listVisitorResponse
	err = json.NewDecoder(listRecorder.Body).Decode(&resp)
	require.NoError(t, err)
	require.LessOrEqual(t, len(resp), 5)
	require.Contains(t, resp[0].Ip, "127.0.0")
	require.Equal(t, resp[0].OriginalURL, data)
	require.Equal(t, resp[0].ShortenURL, shortenURL.ShortenURL)

	// Clean up database
	server.queries.DeleteVisitor(context.Background(), db.DeleteVisitorParams{
		Ip:          resp[0].Ip,
		UrlID:       service.DecodeBase62(code),
		TimeVisited: resp[0].TimeVisited,
	})
	server.queries.DeleteURL(context.Background(), data)
}
