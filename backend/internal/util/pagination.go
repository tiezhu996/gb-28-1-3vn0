package util

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PageParams 分页参数。
type PageParams struct {
	Page     int64
	PageSize int64
}

// GetPageParams 从 query 中解析 page / page_size，带默认值与上限。
func GetPageParams(c *gin.Context, defaultSize int64) PageParams {
	page := parseInt64(c.Query("page"), 1)
	pageSize := parseInt64(c.Query("page_size"), defaultSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultSize
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return PageParams{Page: page, PageSize: pageSize}
}

// TotalPages 计算总页数。
func TotalPages(total, pageSize int64) int64 {
	if pageSize <= 0 {
		return 0
	}
	return int64(math.Ceil(float64(total) / float64(pageSize)))
}

func parseInt64(s string, def int64) int64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return v
}
