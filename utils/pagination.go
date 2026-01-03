package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

// PaginationResponse holds pagination metadata
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// GetPaginationParams extracts pagination parameters from query string
func GetPaginationParams(c *gin.Context, defaultPageSize int, maxPageSize int) PaginationParams {
	page := 1
	pageSize := defaultPageSize

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			if ps > maxPageSize {
				pageSize = maxPageSize
			} else {
				pageSize = ps
			}
		}
	}

	offset := (page - 1) * pageSize

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// CalculateTotalPages calculates total pages based on total records and page size
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize == 0 {
		return 0
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	return totalPages
}

// BuildPaginationResponse builds pagination response metadata
func BuildPaginationResponse(page, pageSize int, total int64) PaginationResponse {
	return PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: CalculateTotalPages(total, pageSize),
	}
}
