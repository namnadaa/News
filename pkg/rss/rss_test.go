package rss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"news/pkg/storage"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	const rssXML = `<rss version="2.0">
    <channel>
        <title>Demo</title>
        <item>
            <title>Test Title #1</title>
            <guid isPermaLink="true">https://example.com</guid>
            <link>
			https://example.ru/ru/articles/198744/=rss
			</link>
            <description>
            <![CDATA[Test description #1]]>
            </description>
            <pubDate>Wed, 08 Dec 2001 01:02:03 GMT</pubDate>
            <dc:creator>
            <![CDATA[ itcaat ]]>
            </dc:creator>
            <category>category_test</category>
        </item>
        <item>
            <title>Test Title #2</title>
            <link>
			https://example.ru/ru/articles/108175/=rss
			</link>
            <description>
            <![CDATA[Test description #2]]>
            </description>
            <pubDate>Mon Jan 02 15:04:05 -0700 2006</pubDate>
        </item>
        <item>
            <title>Test Title #3</title>
            <link>
			https://example.ru/com/articles/1974014/=rss
			</link>
            <description>
            <![CDATA[Test description #3]]>
            </description>
            <pubDate></pubDate>
        </item>
    </channel>
    </rss>`

	post := []storage.Post{
		{
			Title:   "Test Title #1",
			Content: "Test description #1",
			PubTime: time.Date(2001, 12, 8, 01, 02, 03, 0, time.UTC),
			Link:    "https://example.ru/ru/articles/198744/=rss",
		},
		{
			Title:   "Test Title #2",
			Content: "Test description #2",
			PubTime: time.Date(2006, 1, 2, 22, 04, 05, 0, time.UTC),
			Link:    "https://example.ru/ru/articles/108175/=rss",
		},
		{
			Title:   "Test Title #3",
			Content: "Test description #3",
			PubTime: time.Date(0001, 1, 1, 00, 00, 00, 00, time.UTC),
			Link:    "https://example.ru/com/articles/1974014/=rss",
		},
	}

	testOK := struct {
		want    []storage.Post
		wantErr bool
	}{
		want:    post,
		wantErr: false,
	}

	t.Run("Valid RSS", func(t *testing.T) {
		tsOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(rssXML))
		}))
		defer tsOK.Close()

		client := &http.Client{Timeout: 10 * time.Second}
		ctx := context.Background()

		got, err := Parse(ctx, client, tsOK.URL)
		if (err != nil) != testOK.wantErr {
			t.Errorf("Parse() error = %v, wantErr %v", err, testOK.wantErr)
			return
		}
		if !reflect.DeepEqual(got, testOK.want) {
			t.Errorf("Parse() = %v, want %v", got, testOK.want)
		}
	})
	t.Run("Non-2xx status code", func(t *testing.T) {
		tsErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			longBody := strings.Repeat("err", 1000)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(longBody))
		}))
		defer tsErr.Close()

		client := &http.Client{Timeout: 10 * time.Second}
		ctx := context.Background()

		wantErr := true
		_, err := Parse(ctx, client, tsErr.URL)
		if (err != nil) != wantErr {
			t.Errorf("Parse() error = %v, wantErr %v", err, wantErr)
			return
		}

	})
}
