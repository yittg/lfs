package lfs

import (
	"errors"
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

	if err := c.PreparePath(); err != nil {
		log.Fatal(err, "Failed to prepare workspace directory", "path", c.Path)
	}
}

func (c *Configuration) PreparePath() error {
	if path, err := filepath.Abs(c.Path); err != nil {
		return err
	} else {
		c.Path = filepath.Clean(path)
	}
	if c.Path == "/" {
		return errors.New("should not use root dir")
	}
	if err := CreateDir(c.Path); err != nil {
		return err
	}
	return nil
}
