package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Gyscos/benchbase"
)

type BenchTableTitle struct {
	Focus  string
	Title  string
	Width  int
	Height int
}

type TimeResult struct {
	Total bool
	Time  string
}

type BenchTableRows struct {
	TableID int
	Rows    []BenchTableRow
}

type BenchTableRow struct {
	Id      int
	Host    string
	Rev     string
	Threads string
	Date    string

	Group int
	Times []TimeResult
}

func makeBenchGroup(tree *TimeTree, benchlist []benchbase.Benchmark, ignoredSpecs benchbase.Configuration, group int) BenchTableRows {
	result := BenchTableRows{
		Rows: make([]BenchTableRow, len(benchlist))}
	for i, bench := range benchlist {
		row := makeBenchRow(tree, bench, ignoredSpecs)
		row.Id = i
		row.Group = group % 7
		result.Rows[i] = row
	}

	return result
}

func makeBenchRow(tree *TimeTree, bench benchbase.Benchmark, ignoredSpecs benchbase.Configuration) BenchTableRow {
	var result BenchTableRow

	result.Date = bench.Date.Format("2006-01-02")

	if bench.Conf["Host"] != ignoredSpecs["Host"] {
		result.Host = bench.Conf["Host"]
	}

	if bench.Conf["Rev"] != ignoredSpecs["Rev"] {
		_, err := strconv.Atoi(bench.Conf["Rev"])
		if err == nil {
			result.Rev = bench.Conf["Rev"]
		}
	}

	if bench.Conf["Threads"] != ignoredSpecs["Threads"] {
		result.Threads = bench.Conf["Threads"]
	}

	addBenchResults(&result.Times, tree, bench.Result)

	return result
}

func addBenchResults(times *[]TimeResult, tree *TimeTree, results benchbase.Result) {
	for _, child := range tree.children {
		if len(child.children) == 0 {
			time := fmt.Sprintf("%.2f", results[child.prefix+child.name])
			*times = append(*times, TimeResult{child.name == "total", time})
		} else {
			addBenchResults(times, child, results)
		}
	}
}

func makeTimeLabels(tree *TimeTree, conf benchbase.Configuration, specName string) [][]BenchTableTitle {

	var result [][]BenchTableTitle

	depth := 0
	heap := tree.children
	for len(heap) != 0 {
		depth++

		var titles []BenchTableTitle
		var newHeap []*TimeTree

		if depth == 1 {
			// Special case: we insert host
			titles = append(titles, BenchTableTitle{
				Title:  specName,
				Width:  1,
				Height: tree.depth,
			})
		}

		for _, node := range heap {
			height := 1
			if node.depth == 1 {
				height = tree.depth - depth
			}
			focus := node.prefix + node.name
			if node.name == "total" {
				focus = strings.TrimSuffix(node.prefix, ".")
			}
			title := BenchTableTitle{
				Title:  node.name,
				Focus:  focus,
				Width:  node.width,
				Height: height,
			}

			titles = append(titles, title)

			for _, child := range node.children {
				newHeap = append(newHeap, child)
			}
		}

		result = append(result, titles)

		heap = newHeap
	}

	return result
}
