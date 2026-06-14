package toutiao

import (
	"time"
)

// Article is the domain record emitted for feed and hot commands.
type Article struct {
	Rank     int    `json:"rank"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Genre    string `json:"genre"`
	Duration string `json:"duration"`
	Abstract string `json:"abstract"`
	Date     string `json:"date"`
	URL      string `json:"url"`
}

// ─── wire types (unexported) ──────────────────────────────────────────────────

type wireResponse struct {
	HasMore bool          `json:"has_more"`
	Message string        `json:"message"`
	Data    []wireArticle `json:"data"`
}

type wireArticle struct {
	Title          string `json:"title"`
	Abstract       string `json:"abstract"`
	Source         string `json:"source"`
	SourceURL      string `json:"source_url"`
	ArticleGenre   string `json:"article_genre"`
	ChineseTag     string `json:"chinese_tag"`
	HasVideo       bool   `json:"has_video"`
	VideoDuration  string `json:"video_duration_str"`
	VideoPlayCount int    `json:"video_play_count"`
	BehotTime      int64  `json:"behot_time"`
	ItemID         string `json:"item_id"`
}

// ─── conversion ───────────────────────────────────────────────────────────────

// wireToArticle converts a raw feed item to the domain Article type.
func wireToArticle(w wireArticle, rank int) Article {
	// URL always uses the production host, not cfg.BaseURL.
	url := "https://www.toutiao.com/group/" + w.ItemID + "/"

	date := ""
	if w.BehotTime > 0 {
		date = time.Unix(w.BehotTime, 0).UTC().Format("2006-01-02")
	}

	abstract := truncateRunes(w.Abstract, 100)

	return Article{
		Rank:     rank,
		Title:    w.Title,
		Source:   w.Source,
		Genre:    w.ArticleGenre,
		Duration: w.VideoDuration,
		Abstract: abstract,
		Date:     date,
		URL:      url,
	}
}

// truncateRunes truncates s to at most n runes, appending "..." if truncated.
func truncateRunes(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n]) + "..."
}
