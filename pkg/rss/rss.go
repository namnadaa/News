package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"news/pkg/storage"
	"time"

	strip "github.com/grokify/html-strip-tags-go"
)

// Internal structures for unpacking RSS-XML.
type feed struct {
	XMLName xml.Name `xml:"rss"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []item `xml:"item"`
}

type item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
}

// Parse downloads RSS from the URL, decodes the XML, and returns a Post slice.
func Parse(url string) ([]storage.Post, error) {
	client := &http.Client{Timeout: time.Duration(time.Second * 10)}
	res, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http GET failed: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %v", err)
	}

	if res.StatusCode > 299 {
		limit := body
		if len(limit) > 2048 {
			limit = limit[:2048]
		}
		return nil, fmt.Errorf("response failed with status code: %d and body: %s", res.StatusCode, limit)
	}

	var f feed
	err = xml.Unmarshal(body, &f)
	if err != nil {
		return nil, fmt.Errorf("XML unmarshal failed: %v", err)
	}

	layouts := []string{
		time.Layout,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
	}

	var data []storage.Post
	for _, itemNode := range f.Channel.Items {
		var p storage.Post
		p.Title = itemNode.Title
		p.Content = strip.StripTags(itemNode.Description)
		p.Link = itemNode.Link

		var parsed time.Time
		var parsErr error
		for _, l := range layouts {
			t, err := time.Parse(l, itemNode.PubDate)
			if err != nil {
				parsErr = err
			}
			if err == nil {
				parsed = t
				break
			}
		}

		if parsed.IsZero() {
			slog.Warn("Parse: pubDate parse failed", "pubDate", itemNode.PubDate, "err", parsErr)
		}
		p.PubTime = parsed

		data = append(data, p)
	}

	return data, nil
}
