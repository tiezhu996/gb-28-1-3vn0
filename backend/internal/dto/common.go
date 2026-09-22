// Package dto 定义接口出入参结构。
package dto

// PageResult 统一分页返回结构。
type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int64       `json:"page"`
	PageSize int64       `json:"page_size"`
	TotalPage int64      `json:"total_page"`
}
