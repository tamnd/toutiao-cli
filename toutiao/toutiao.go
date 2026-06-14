// Package toutiao is the library behind the toutiao command: the HTTP client,
// request shaping, and the typed data models for the Toutiao (今日头条) news feed.
//
// The client issues GET requests to the public Toutiao PC feed API at
// https://www.toutiao.com/api/pc/feed/. No authentication is required. It sets
// a real User-Agent plus the mandatory Referer header, paces requests, and
// retries transient 429/5xx errors with exponential backoff.
package toutiao

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultUserAgent is the browser-like UA sent with every request.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// referer is hardcoded and not derived from cfg.BaseURL — the API requires it.
const referer = "https://www.toutiao.com/"

// categories maps friendly aliases to the API category parameter values.
var categories = map[string]string{
	"hot":   "news_hot",
	"video": "video",
	"all":   "__all__",
}

// Config holds constructor parameters for Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults for talking to toutiao.com.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://www.toutiao.com",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client talks to the Toutiao feed API.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// Feed fetches the Toutiao feed for the given category alias (hot, video, all)
// and returns up to limit articles. If limit <= 0, all returned articles are
// included. category may also be a raw API category string (e.g. "news_hot").
func (c *Client) Feed(ctx context.Context, category string, limit int) ([]Article, error) {
	cat, ok := categories[category]
	if !ok {
		cat = category
	}

	url := fmt.Sprintf("%s/api/pc/feed/?category=%s&utm_source=toutiao&wid=1&max_behot_time=0",
		c.cfg.BaseURL, cat)

	raw, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}

	var resp wireResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode feed: %w", err)
	}

	data := resp.Data
	if limit > 0 && len(data) > limit {
		data = data[:limit]
	}

	out := make([]Article, 0, len(data))
	for i, item := range data {
		out = append(out, wireToArticle(item, i+1))
	}
	return out, nil
}

// get issues a GET request with retry logic.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, url)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Referer", referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
