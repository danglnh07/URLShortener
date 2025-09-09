package api

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestRateLimiterStress(t *testing.T) {
	// Handler always responds OK if limiter allows
	handler := server.RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	svr := httptest.NewServer(handler)
	defer svr.Close()

	client := &http.Client{}
	var wg sync.WaitGroup
	success := 0
	tooMany := 0
	total := 50

	for range total {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Get(svr.URL)
			if err != nil {
				t.Errorf("request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				success++
			case http.StatusTooManyRequests:
				tooMany++
			default:
				t.Errorf("unexpected status: %d", resp.StatusCode)
			}
		}()
	}

	wg.Wait()

	t.Logf("Success=%d, TooMany=%d", success, tooMany)

	if success > 5 {
		t.Errorf("Limiter failed: expected at most 5 OK responses, got %d", success)
	}
}
