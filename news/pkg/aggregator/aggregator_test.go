package aggregator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"news/pkg/storage"
	"testing"
	"time"
)

const rssXML = `<rss version="2.0">
    <channel>
        <title>Demo</title>
        <item>
            <title>Test Title #1</title>
            <link>
			https://example.ru/ru/articles/108175/=rss
			</link>
            <description>
            <![CDATA[Test description #1]]>
            </description>
            <pubDate>Mon Jan 02 15:04:05 -0700 2006</pubDate>
        </item>
        <item>
            <title>Test Title #2</title>
            <link>
			https://example.ru/com/articles/1974014/=rss
			</link>
            <description>
            <![CDATA[Test description #2]]>
            </description>
            <pubDate></pubDate>
        </item>
    </channel>
    </rss>`

func TestAggregator_runOnce(t *testing.T) {
	t.Run("Valid Test", func(t *testing.T) {
		tsOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(rssXML))
		}))
		defer tsOK.Close()

		ag := New([]string{tsOK.URL}, 0)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		postCh := make(chan []storage.Post)
		errsCh := make(chan error)

		go ag.runOnce(ctx, postCh, errsCh)

		select {
		case got := <-postCh:
			if len(got) != 2 {
				t.Fatalf("got %d - want 2 posts", len(got))
			}

			if got[0].Title != "Test Title #1" || got[1].Title != "Test Title #2" {
				t.Fatalf("unexpected titles: %v", got)
			}
		case err := <-errsCh:
			t.Fatalf("unexpected error: %v", err)
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for posts")
		}
	})

	t.Run("Error Status", func(t *testing.T) {
		tsOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(rssXML))
		}))
		defer tsOK.Close()

		ag := New([]string{tsOK.URL}, 0)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		postCh := make(chan []storage.Post)
		errsCh := make(chan error)

		go ag.runOnce(ctx, postCh, errsCh)

		select {
		case <-postCh:
			t.Fatalf("unexpected posts on error")
		case err := <-errsCh:
			if err == nil {
				t.Fatalf("got nil - want error")
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for error")
		}
	})
}

func TestAggregator_Run(t *testing.T) {

	t.Run("Valid Period", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(rssXML))
		}))
		defer ts.Close()

		ag := New([]string{ts.URL}, 20*time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		postCh := make(chan []storage.Post)
		errsCh := make(chan error)

		go ag.Run(ctx, postCh, errsCh)

		select {
		case got := <-postCh:
			if len(got) != 2 {
				t.Fatalf("got %d - want 2 posts", len(got))
			}

			if got[0].Title != "Test Title #1" || got[1].Title != "Test Title #2" {
				t.Fatalf("unexpected titles: %v", got)
			}
		case err := <-errsCh:
			t.Fatalf("unexpected error: %v", err)
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for posts")
		}
	})

	t.Run("Period less 0", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(rssXML))
		}))
		defer ts.Close()

		ag := New([]string{ts.URL}, 0)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		postCh := make(chan []storage.Post)
		errsCh := make(chan error)

		go ag.Run(ctx, postCh, errsCh)

		select {
		case got := <-postCh:
			if len(got) != 2 {
				t.Fatalf("got %d - want 2 posts", len(got))
			}

			if got[0].Title != "Test Title #1" || got[1].Title != "Test Title #2" {
				t.Fatalf("unexpected titles: %v", got)
			}
		case err := <-errsCh:
			t.Fatalf("unexpected error: %v", err)
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for posts")
		}
	})
}
