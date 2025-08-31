package response

import (
	"math"
	"net/url"
	"strconv"
)

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
	NextPage   *int  `json:"next_page,omitempty"`
	PrevPage   *int  `json:"prev_page,omitempty"`
}

type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

func ParsePaginationFromQuery(values url.Values) PaginationParams {
	page, err := strconv.Atoi(values.Get("page"))
	if err != nil || page < 1 {
		page = defaultPage
	}

	limit, err := strconv.Atoi(values.Get("limit"))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

func CreatePaginationMeta(params PaginationParams, total int64) PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))
	hasNext := params.Page < totalPages
	hasPrev := params.Page > 1

	meta := PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}

	if hasNext {
		nextPage := params.Page + 1
		meta.NextPage = &nextPage
	}

	if hasPrev {
		prevPage := params.Page - 1
		meta.PrevPage = &prevPage
	}

	return meta
}

func (r *Response) WithPagination(params PaginationParams, total int64) *Response {
	meta := CreatePaginationMeta(params, total)
	r.PaginationData = &meta
	return r
}
