package steamapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestResolveVanityURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ISteamUser/ResolveVanityURL/v1/" {
			t.Errorf("Expected path /ISteamUser/ResolveVanityURL/v1/, got %s", r.URL.Path)
		}
		_, _ = fmt.Fprintln(w, `{"response": {"steamid": "76561198000000000", "success": 1}}`)
	}))
	defer ts.Close()

	// Override baseURL for testing?
	// The client uses a const baseURL. I need to make it configurable or use a transport interceptor.
	// Since I can't change the const, I'll use a custom transport that redirects requests to the test server.
}

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

func TestClient_ResolveVanityURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response": {"steamid": "76561198000000000", "success": 1}}`))
	}))
	defer ts.Close()

	c, err := NewClient("test-key", time.Minute, time.Minute, time.Second)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	// Clean up cache
	defer func() { _ = os.RemoveAll(c.gamesCache.DirPath()) }()

	// Inject test transport
	c.httpClient.Transport = &TestTransport{TargetURL: ts.URL}

	sid, err := c.ResolveVanityURL(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("ResolveVanityURL error: %v", err)
	}
	if sid != "76561198000000000" {
		t.Errorf("got %s, want 76561198000000000", sid)
	}
}

func TestClient_GetOwnedGames(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"response": {
				"game_count": 1,
				"games": [
					{
						"appid": 10,
						"name": "Counter-Strike",
						"playtime_forever": 0,
						"img_icon_url": "test"
					}
				]
			}
		}`))
	}))
	defer ts.Close()

	c, err := NewClient("test-key", time.Minute, time.Minute, time.Second)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	defer func() { _ = os.RemoveAll(c.gamesCache.DirPath()) }()

	c.httpClient.Transport = &TestTransport{TargetURL: ts.URL}

	games, err := c.GetOwnedGames(context.Background(), "76561198000000000", false)
	if err != nil {
		t.Fatalf("GetOwnedGames error: %v", err)
	}
	if len(games) != 1 {
		t.Errorf("got %d games, want 1", len(games))
	}
	if games[0].Name != "Counter-Strike" {
		t.Errorf("got %s, want Counter-Strike", games[0].Name)
	}
}

func TestClient_GetAppDetails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/appdetails" {
			t.Errorf("Expected path /api/appdetails, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"10": {
				"success": true,
				"data": {
					"name": "Counter-Strike",
					"short_description": "Action game"
				}
			}
		}`))
	}))
	defer ts.Close()

	c, err := NewClient("test-key", time.Minute, time.Minute, time.Second)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	// Clean up cache
	defer func() { _ = os.RemoveAll(c.gamesCache.Dir) }()
	defer func() { _ = os.RemoveAll(c.vanityCache.Dir) }()

	c.httpClient.Transport = &TestTransport{TargetURL: ts.URL}

	details, err := c.GetAppDetails(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetAppDetails error: %v", err)
	}

	entry, ok := (*details)["10"]
	if !ok {
		t.Fatal("Expected details for app 10")
	}
	if !entry.Success {
		t.Error("Expected success true")
	}
	if entry.Data.Name != "Counter-Strike" {
		t.Errorf("Expected name Counter-Strike, got %s", entry.Data.Name)
	}
}
