package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port      int
	ExportDir string
}

func Load() Config {
	p := 8080
	if v, e := strconv.Atoi(os.Getenv("PORT")); e == nil && v > 0 && v < 65536 {
		p = v
	}
	exportDir := os.Getenv("EXPORT_DIR")
	return Config{Port: p, ExportDir: exportDir}
}
func (c Config) Address() string { return fmt.Sprintf(":%d", c.Port) }
