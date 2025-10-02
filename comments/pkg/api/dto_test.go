package api

import (
	"comments/pkg/storage"
	"reflect"
	"testing"
	"time"
)

func Test_toDTOs(t *testing.T) {
	type args struct {
		p []storage.Comment
	}
	tests := []struct {
		name string
		args args
		want []commentDTO
	}{
		{
			name: "Valid Test",
			args: args{
				p: []storage.Comment{
					{
						ID:        "1",
						NewsID:    "123",
						ParentID:  "2",
						Author:    "Alex",
						Content:   "Content 1",
						CreatedAt: time.Date(2005, 10, 12, 13, 15, 0, 0, time.UTC),
					},
					{
						ID:        "2",
						NewsID:    "321",
						ParentID:  "4",
						Author:    "Bob",
						Content:   "Content 2",
						CreatedAt: time.Date(2006, 9, 11, 12, 14, 0, 0, time.UTC),
					},
				},
			},
			want: []commentDTO{
				{
					ID:        "1",
					NewsID:    "123",
					ParentID:  "2",
					Author:    "Alex",
					Content:   "Content 1",
					CreatedAt: 1129122900,
				},
				{
					ID:        "2",
					NewsID:    "321",
					ParentID:  "4",
					Author:    "Bob",
					Content:   "Content 2",
					CreatedAt: 1157976840,
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
