package lfs

import (
	"path/filepath"

	log "github.com/yittg/golog"
)

type Configuration struct {
	BindAddr  string
	ServePort int

	Path string

	UploadFilters []FilterFunc
	FetchFilters  []FilterFunc
	DeleteFilters []FilterFunc
}

func (c *Configuration) SetDefaults() {
	if c.ServePort <= 0 || c.ServePort > 65535 {
		c.ServePort = 8080
	}

	if path, err := filepath.Abs(c.Path); err != nil {
		log.Fatal(err, "Failed to resolve absolute workspace directory", "path", c.Path)
	} else {
		c.Path = path
	}
}
