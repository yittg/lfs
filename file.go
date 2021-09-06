package lfs

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/yittg/golog"
)

type FileHandler struct {
	PathResolver func(string) (string, error)
	PathClearer  func(string) error
}

func (f *FileHandler) CreateFileFilter(ctx *FilterContext) error {
	if ctx.FileContent == nil {
		return os.ErrInvalid
	}

	path, err := f.PathResolver(ctx.FilePath)
	if err != nil {
		return err
	}

	if err := NewStater().NotExistOk().ExpectRegularFile().Stat(path); err != nil {
		return err
	}

	file, err := createFile(path)
	if err != nil {
		log.Error(err, "Failed to create file", "file", path)
		return err
	}
	defer file.Close()
	if _, err = io.Copy(file, ctx.FileContent); err != nil {
		log.Info("Failed to upload file", "err", err)
		return err
	}
	return nil
}

func (f *FileHandler) LoadFile(ctx *FilterContext) error {
	path, err := f.PathResolver(ctx.FilePath)
	if err != nil {
		return err
	}

	if err := NewStater().ExpectRegularFile().Stat(path); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	ctx.FileContent = file
	return nil
}

func (f *FileHandler) DeleteFileFilter(ctx *FilterContext) error {
	path, err := f.PathResolver(ctx.FilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err = f.PathClearer(filepath.Dir(path)); err != nil {
			log.Error(err, "Failed to clear empty dir", "path", path)
		}
	}()
	if err := NewStater().NotExistOk().ExpectRegularFile().Stat(path); err != nil {
		return err
	}

	if err = os.Remove(path); err != nil {
		log.Error(err, "Failed to remove file", "path", path)
		return err
	}
	return nil
}

type Stater struct {
	notExistOk bool
	modFunc    func(os.FileMode) bool
}

func NewStater() *Stater {
	return &Stater{}
}

func (st *Stater) NotExistOk() *Stater {
	st.notExistOk = true
	return st
}

func (st *Stater) ExpectRegularFile() *Stater {
	st.modFunc = func(mode os.FileMode) bool {
		return mode.IsRegular()
	}
	return st
}

func (st *Stater) ExpectDir() *Stater {
	st.modFunc = func(mode os.FileMode) bool {
		return mode.IsDir()
	}
	return st
}

func (st *Stater) Stat(f string) error {
	if dirInfo, err := os.Stat(f); err != nil {
		if os.IsNotExist(err) && st.notExistOk {
			return nil
		}
		return err
	} else {
		if !st.modFunc(dirInfo.Mode()) {
			return os.ErrInvalid
		}
	}
	return nil
}

func FileFilterWrapper(fn func(ctx *FilterContext) error) FilterFunc {
	return func(ctx *FilterContext, next Filter) {
		if err := fn(ctx); err != nil {
			if os.IsNotExist(err) {
				ctx.ResponseCode = http.StatusNotFound
			} else if err == os.ErrInvalid {
				ctx.ResponseCode = http.StatusBadRequest
			} else {
				ctx.ResponseCode = http.StatusInternalServerError
			}
		}
	}
}

func createFile(fullPath string) (*os.File, error) {
	dir := filepath.Dir(fullPath)
	if err := NewStater().NotExistOk().ExpectDir().Stat(dir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}
	return os.Create(fullPath)
}

func SubOf(parent, path string) bool {
	return strings.HasPrefix(path, parent) &&
		path != parent &&
		path[len(parent)] == os.PathSeparator
}

func resolvePath(parent, path string) (string, error) {
	path = filepath.Join(parent, filepath.FromSlash(path))
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if !SubOf(parent, resolved) {
		return "", os.ErrInvalid
	}
	return resolved, nil
}

func removeEmptyDir(parent, path string) error {
	if !SubOf(parent, path) {
		return os.ErrInvalid
	}
	for path != parent {
		if err := NewStater().ExpectDir().Stat(path); err != nil {
			if os.IsNotExist(err) {
				path = filepath.Dir(path)
				continue
			}
			return err
		}
		if empty, err := isEmptyDir(path); err != nil || !empty {
			return err
		}
		if err := os.Remove(path); err != nil {
			log.Error(err, "Failed to remove file", "path", path)
			return err
		}
		path = filepath.Dir(path)
	}
	return nil
}

func isEmptyDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
