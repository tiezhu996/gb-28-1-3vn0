// Package repository 定义仓储层接口与 MongoDB 实现。
package repository

import "errors"

// 仓储层哨兵错误，service 层使用 errors.Is 判断。
var (
	ErrNotFound        = errors.New("repository: not found")
	ErrConflict        = errors.New("repository: conflict")
	ErrVersionConflict = errors.New("repository: draft version conflict")
)
