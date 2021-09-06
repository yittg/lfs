package lfs

import (
	"io"
	"net/http"
)

type FilterContext struct {
	Request *http.Request

	FilePath    string
	FileContent io.Reader

	ResponseCode    int
	ResponseHeader  http.Header
	ResponseContent io.Reader
}

type Filter interface {
	Do(ctx *FilterContext)
}

type FilterFunc func(ctx *FilterContext, next Filter)

type ChainedFilter struct {
	Next       Filter
	FilterFunc FilterFunc
}

func NewChainedFilter() ChainedFilter {
	return ChainedFilter{
		FilterFunc: func(*FilterContext, Filter) {},
	}
}

func (s ChainedFilter) Do(ctx *FilterContext) {
	s.FilterFunc(ctx, s.Next)
}

func (s ChainedFilter) Chain(filterFunc FilterFunc) ChainedFilter {
	return ChainedFilter{
		Next:       s,
		FilterFunc: filterFunc,
	}
}

func BuildFilters(filters []FilterFunc) Filter {
	filter := NewChainedFilter()
	for i := len(filters) - 1; i >= 0; i-- {
		filter = filter.Chain(filters[i])
	}
	return filter
}
