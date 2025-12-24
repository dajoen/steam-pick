package storeapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Transport to redirect requests to test server
type TestTransport struct {
	TargetURL string
}

func (t *TestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u, _ := req.URL.Parse(t.TargetURL)
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	return http.DefaultTransport.RoundTrip(req)
}

func TestClient_IsTurnBased(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"10": {
				"success": true,
				"data": {
					"name": "Turn Based Game",
					"genres": [{"id": "1", "description": "Turn-based Strategy"}]
				}
			},
			"20": {
				"success": true,
				"data": {
					"name": "Action Game",
					"genres": [{"id": "1", "description": "Action"}]
				}
			}
		}`))
	}))
	defer ts.Close()

	c := NewClient(time.Second)
	c.httpClient.Transport = &TestTransport{TargetURL: ts.URL}

	// Test Turn Based
	isTurn, err := c.IsTurnBased(context.Background(), 10, "US")
	if err != nil {
		t.Errorf("IsTurnBased(10) error: %v", err)
	}
	if !isTurn {
		t.Errorf("IsTurnBased(10) = false, want true")
	}

	// Test Not Turn Based
	isTurn, err = c.IsTurnBased(context.Background(), 20, "US")
	if err != nil {
		t.Errorf("IsTurnBased(20) error: %v", err)
	}
	if isTurn {
		t.Errorf("IsTurnBased(20) = true, want false")
	}
}
