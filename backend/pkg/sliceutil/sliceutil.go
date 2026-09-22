// Package sliceutil 提供与业务无关的切片工具（可复用库）。
package sliceutil

// Contains 判断切片是否包含目标元素。
func Contains[T comparable](items []T, target T) bool {
	for _, it := range items {
		if it == target {
			return true
		}
	}
	return false
}

// Dedupe 去重（保持原顺序）。
func Dedupe[T comparable](items []T) []T {
	seen := make(map[T]struct{}, len(items))
	out := make([]T, 0, len(items))
	for _, it := range items {
		if _, ok := seen[it]; ok {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
	}
	return out
}

// Map 对切片逐元素转换。
func Map[T, U any](items []T, fn func(T) U) []U {
	out := make([]U, 0, len(items))
	for _, it := range items {
		out = append(out, fn(it))
	}
	return out
}
