package toutiao_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/toutiao-cli/toutiao"
)

func TestFeed(t *testing.T) {
	payload := map[string]any{
		"has_more": true,
		"data": []map[string]any{
			{
				"title":         "Test Article",
				"abstract":      "Test abstract",
				"source":        "Test Source",
				"article_genre": "article",
				"behot_time":    int64(1700000000),
				"item_id":       "7312345678901234567",
			},
		},
	}
	b, _ := json.Marshal(payload)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write(b)
	}))
	defer srv.Close()

	cfg := toutiao.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := toutiao.NewClient(cfg)
	articles, err := c.Feed(context.Background(), "hot", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 {
		t.Fatalf("got %d articles, want 1", len(articles))
	}
	if articles[0].Title != "Test Article" {
		t.Errorf("title = %q, want %q", articles[0].Title, "Test Article")
	}
	if articles[0].Rank != 1 {
		t.Errorf("rank = %d, want 1", articles[0].Rank)
	}
}

func TestFeedLimit(t *testing.T) {
	items := make([]map[string]any, 5)
	for i := range items {
		items[i] = map[string]any{"title": "t", "item_id": "123", "behot_time": int64(1700000000)}
	}
	payload := map[string]any{"data": items}
	b, _ := json.Marshal(payload)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(b)
	}))
	defer srv.Close()

	cfg := toutiao.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := toutiao.NewClient(cfg)
	articles, err := c.Feed(context.Background(), "all", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 3 {
		t.Fatalf("got %d articles, want 3 (limit applied)", len(articles))
	}
}
