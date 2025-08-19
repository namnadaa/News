package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	RSS           []string `json:"rss"`
	RequestPeriod int      `json:"request_period"`
}

func (c *Config) Load(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config reading error: %v", err)
	}

	err = json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("deserialization error: %v", err)
	}

	double := make(map[string]struct{}, len(c.RSS))
	out := make([]string, 0, len(c.RSS))
	for _, raw := range c.RSS {
		s := strings.TrimSpace(raw)

		if _, ok := double[s]; !ok {
			double[s] = struct{}{}
			out = append(out, s)
		}
	}

	if len(out) == 0 {
		return fmt.Errorf("rss list is empty")
	}

	if c.RequestPeriod <= 0 {
		c.RequestPeriod = 5
	}

	c.RSS = out
	return nil
}
