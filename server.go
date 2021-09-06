package lfs

import (
	"fmt"
	"io"
	"net/http"

	log "github.com/yittg/golog"
)

type Server struct {
	cfg *Configuration

	fetchFilter  Filter
	uploadFilter Filter
	deleteFilter Filter
}

func NewServer(cfg *Configuration) *Server {
	cfg.SetDefaults()

	fileHandler := &FileHandler{
		PathResolver: func(path string) (string, error) {
			return resolvePath(cfg.Path, path)
		},
		PathClearer: func(path string) error {
			return removeEmptyDir(cfg.Path, path)
		},
	}

	fetchFilter := BuildFilters(append(cfg.FetchFilters, FileFilterWrapper(fileHandler.LoadFile)))
	uploadFilter := BuildFilters(append(cfg.UploadFilters, FileFilterWrapper(fileHandler.CreateFileFilter)))
	deleteFilter := BuildFilters(append(cfg.DeleteFilters, FileFilterWrapper(fileHandler.DeleteFileFilter)))

	return &Server{
		cfg:          cfg,
		fetchFilter:  fetchFilter,
		uploadFilter: uploadFilter,
		deleteFilter: deleteFilter,
	}
}

func (s *Server) Run() {
	if err := NewStater().ExpectDir().Stat(s.cfg.Path); err != nil {
		log.Fatal(err, "Path is invalid", "path", s.cfg.Path)
	}
	log.Info("Starting file server", "path", s.cfg.Path)

	http.HandleFunc("/", s.handleRequest)
	addr := fmt.Sprintf("%s:%d", s.cfg.BindAddr, s.cfg.ServePort)
	err := http.ListenAndServe(addr, nil)
	log.Error(err, "Failed to listen", "addr", addr)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	log.V(5).Info("Receive request", "method", r.Method, "path", r.URL.Path)

	if r.Method == http.MethodGet {
		s.loadFile(w, r)
	} else if r.Method == http.MethodPost {
		s.uploadFile(w, r)
	} else if r.Method == http.MethodDelete {
		s.deleteFile(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) loadFile(w http.ResponseWriter, r *http.Request) {
	ctx := &FilterContext{
		Request:        r,
		FilePath:       r.URL.Path,
		ResponseHeader: w.Header(),
	}
	s.fetchFilter.Do(ctx)
	if ctx.ResponseContent == nil {
		ctx.ResponseContent = ctx.FileContent
	}
	respond(ctx, w)
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := &FilterContext{
		Request:        r,
		FilePath:       r.URL.Path,
		FileContent:    r.Body,
		ResponseHeader: w.Header(),
	}
	s.uploadFilter.Do(ctx)
	respond(ctx, w)
}

func (s *Server) deleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := &FilterContext{
		Request:        r,
		FilePath:       r.URL.Path,
		ResponseHeader: w.Header(),
	}
	s.deleteFilter.Do(ctx)
	respond(ctx, w)
}

func respond(ctx *FilterContext, w http.ResponseWriter) {
	if ctx.ResponseCode != 0 && ctx.ResponseCode != http.StatusOK {
		w.WriteHeader(ctx.ResponseCode)
		return
	}
	if ctx.ResponseContent != nil {
		io.Copy(w, ctx.ResponseContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if closer, ok := ctx.ResponseContent.(io.Closer); ok {
		closer.Close()
	}
}
