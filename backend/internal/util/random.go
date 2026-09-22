package util

import "math/rand"

// Shuffle 随机打乱切片（用于题目顺序/选项顺序随机化，防作弊）。
func Shuffle[T any](items []T) {
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
}
