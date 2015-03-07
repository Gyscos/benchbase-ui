package main

import (
	"strings"

	"github.com/Gyscos/benchbase"
)

func findBestDepth(result benchbase.Result) int {
	depths := make(map[int]int)
	deepest := 0

	for k, _ := range result {
		depth := strings.Count(k, ".") + 1
		if strings.HasSuffix(k, ".total") {
			depth--
		}
		if depth > deepest {
			deepest = depth
		}
		depths[depth]++
	}

	// Find the deepest value with reasonable width
	best := 0
	width := 0
	for depth := 0; depth <= deepest; depth++ {
		width += depths[depth]
		if width < 17 {
			best = depth
		} else {
			break
		}
	}

	return best
}

func cutOutOfFocus(benchlist []benchbase.Benchmark, focus string) {
	for _, b := range benchlist {
		for k, _ := range b.Result {
			if k != focus && !strings.HasPrefix(k, focus+".") {
				delete(b.Result, k)
			}
		}
	}
}
func truncateDeepResults(benchlist []benchbase.Benchmark, depth int) {
	for _, b := range benchlist {
		var totals []string
		for k, _ := range b.Result {
			c := strings.Count(k, ".")
			if c >= depth {
				if c == depth && strings.HasSuffix(k, ".total") {
					totals = append(totals, k)
				} else {
					delete(b.Result, k)
				}
			}
		}
		for _, k := range totals {
			b.Result[strings.TrimSuffix(k, ".total")] = b.Result[k]
			delete(b.Result, k)
		}
	}
}
