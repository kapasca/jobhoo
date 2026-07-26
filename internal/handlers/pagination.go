package handlers

import (
	"math"
	"net/http"
	"strconv"
)

type Pagination struct {
	Page       int
	PageSize   int
	TotalItems int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	PrevPage   int
	NextPage   int
}

func parsePageParam(r *http.Request, key string) int {
	page, _ := strconv.Atoi(r.URL.Query().Get(key))
	if page < 1 {
		return 1
	}
	return page
}

func buildPagination(totalItems, page, pageSize int) Pagination {
	if pageSize <= 0 {
		pageSize = 10
	}
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	p := Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
	if page > 1 {
		p.HasPrev = true
		p.PrevPage = page - 1
	}
	if page < totalPages {
		p.HasNext = true
		p.NextPage = page + 1
	}
	return p
}

func paginationOffset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}
