package api

import (
	"news/pkg/storage"
	"reflect"
	"testing"
	"time"
)

func Test_toDTOs(t *testing.T) {
	type args struct {
		p []storage.Post
	}
	tests := []struct {
		name string
		args args
		want []postDTO
	}{
		{
			name: "Valid Test",
			args: args{
				p: []storage.Post{
					{
						ID:      1,
						Title:   "Title 1",
						Content: "Content 1",
						PubTime: time.Date(2006, 9, 11, 12, 14, 0, 0, time.UTC),
						Link:    "https://example.ru/ru/articles/108175/=rss",
					},
					{
						ID:      2,
						Title:   "Title 2",
						Content: "Content 2",
						PubTime: time.Date(2005, 10, 12, 13, 15, 0, 0, time.UTC),
						Link:    "https://example.ru/ru/articles/981612/=rss",
					},
				},
			},
			want: []postDTO{
				{
					ID:      1,
					Title:   "Title 1",
					Content: "Content 1",
					PubTime: 1157976840,
					Link:    "https://example.ru/ru/articles/108175/=rss",
				},
				{
					ID:      2,
					Title:   "Title 2",
					Content: "Content 2",
					PubTime: 1129122900,
					Link:    "https://example.ru/ru/articles/981612/=rss",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toDTOs(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDTOs() = %v, want %v", got, tt.want)
			}
		})
	}
}
