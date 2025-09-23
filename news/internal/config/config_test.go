package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	testDir := t.TempDir()

	type fields struct {
		RSS           []string
		RequestPeriod int
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid JSON",
			fields: fields{
				RSS: []string{
					"https://habr.com/ru/rss/hub/go/all/?fl=ru",
					"https://habr.com/ru/rss/best/daily/?fl=ru",
					"https://cprss.s3.amazonaws.com/golangweekly.com.xml"},
				RequestPeriod: 5,
			},
			args:    args{filepath.Join(testDir, "valid.json")},
			wantErr: false,
		},
		{
			name: "Empty JSON",
			fields: fields{
				RSS:           []string{},
				RequestPeriod: 0,
			},
			args:    args{filepath.Join(testDir, "empty.json")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := struct {
				RSS           []string `json:"rss"`
				RequestPeriod int      `json:"request_period"`
			}{
				RSS:           tt.fields.RSS,
				RequestPeriod: tt.fields.RequestPeriod,
			}

			b, err := json.Marshal(testConfig)
			if err != nil {
				t.Fatalf("Marshal test json: %v", err)
			}

			err = os.WriteFile(tt.args.path, b, 0644)
			if err != nil {
				t.Fatalf("Write test file: %v", err)
			}

			c := &Config{}
			if err := c.Load(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Config.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
