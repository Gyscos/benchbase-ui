package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Gyscos/benchbase"
)

type BenchListTableTitle struct {
	Focus  string
	Title  string
	Width  int
	Height int
}

type BenchListRow struct {
	Name  string
	Group int
	Times []TimeResult
}

type TimeResult struct {
	Total bool
	Time  string
}

type BenchListTable struct {
	Category string
	Width    int

	Titles [][]BenchListTableTitle

	BenchList []BenchListRow
}

func MakeListTables(host string, filter string, focus string, depth int) ([]BenchListTable, error) {
	resp, err := http.Get(fmt.Sprintf("%v/list?filter=%v", host, filter))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Error  string
		Result []benchbase.Benchmark
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&data)

	if data.Error != "" {
		return nil, errors.New(data.Error)
	}

	if focus != "" {
		cutOutOfFocus(data.Result, focus+".")
		if depth > 0 {
			depth += 1 + strings.Count(focus, ".")
		}
	}

	if depth < 0 {
		// Find best depth
		depth = findBestDepth(data.Result[0].Result)
	}

	if depth != 0 {
		truncateDeepResults(data.Result, depth)
	}

	groups := groupByConf(data.Result)

	result := make([]BenchListTable, len(groups))
	for i, g := range groups {
		result[i] = makeListTable(g)
	}

	return result, nil
}

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
			if !strings.HasPrefix(k, focus) {
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

func groupByConf(benchlist []benchbase.Benchmark) [][]benchbase.Benchmark {
	m := make(map[string][]benchbase.Benchmark)
	for _, b := range benchlist {
		m[b.Conf["ForceAnalyze"]] = append(m[b.Conf["ForceAnalyze"]], b)
	}

	var result [][]benchbase.Benchmark
	for _, l := range m {
		result = append(result, l)
	}
	return result
}

func makeListTable(benchlist []benchbase.Benchmark) BenchListTable {
	var result BenchListTable

	if len(benchlist) == 0 {
		return result
	}

	tree := makeTimeTree(benchlist[0].Result)
	mergeSingleChilds(tree)
	computeDepthWidth(tree)

	result.Category = categoryName(benchlist[0].Conf)
	result.Titles = makeTimeLabels(tree, benchlist[0].Conf, "Host")
	result.BenchList = makeBenchList(tree, benchlist, benchbase.Configuration{}, 0)
	result.Width = tree.width + 1

	return result
}

type TimeTree struct {
	// Prefix:
	prefix   string
	name     string
	depth    int
	width    int
	children []*TimeTree
}

func getLabels(times benchbase.Result) []string {
	var result []string
	for k, _ := range times {
		result = append(result, k)
	}

	sort.StringSlice(result).Sort()

	return result
}

// From a list of time results, make a hierarchy of the labels
func makeTimeTree(times benchbase.Result) *TimeTree {
	var result TimeTree

	labels := getLabels(times)
	for _, k := range labels {
		tokens := strings.Split(k, ".")
		current := &result
		for _, token := range tokens {
			i := findChild(*current, token)
			if i == -1 {
				child := &TimeTree{name: token}
				if current.name != "" {
					if current.prefix != "" {
						child.prefix = current.prefix
					}
					child.prefix += current.name + "."
				}
				if child.name == "total" {
					current.children = append([]*TimeTree{child}, current.children...)
				} else {
					current.children = append(current.children, child)
				}
				i = len(current.children) - 1
			}
			current = current.children[i]
		}
	}

	return &result
}

func findChild(node TimeTree, name string) int {
	for i, child := range node.children {
		if child.name == name {
			return i
		}
	}
	return -1
}

func mergeSingleChilds(node *TimeTree) {
	if node.name != "" && len(node.children) == 1 {
		node.name += "." + node.children[0].name
		node.children = node.children[0].children
		mergeSingleChilds(node)
	} else {
		for _, child := range node.children {
			mergeSingleChilds(child)
		}
	}
}

func categoryName(conf benchbase.Configuration) string {
	if conf["ForceAnalyze"] == "true" {
		return "Analyze API"
	} else {
		return "Direct API"
	}
}

func makeTimeLabels(tree *TimeTree, conf benchbase.Configuration, specName string) [][]BenchListTableTitle {

	var result [][]BenchListTableTitle

	depth := 0
	heap := tree.children
	for len(heap) != 0 {
		depth++

		var titles []BenchListTableTitle
		var newHeap []*TimeTree

		if depth == 1 {
			// Special case: we insert host
			titles = append(titles, BenchListTableTitle{
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
			title := BenchListTableTitle{
				Title:  node.name,
				Focus:  node.prefix + node.name,
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

func computeDepthWidth(node *TimeTree) {
	if len(node.children) == 0 {
		node.width = 1
		node.depth = 1
		return
	}

	for _, child := range node.children {
		computeDepthWidth(child)
		if child.depth+1 > node.depth {
			node.depth = child.depth + 1
		}
		node.width += child.width
	}
}

func makeBenchList(tree *TimeTree, benchlist []benchbase.Benchmark, ignoredSpecs benchbase.Configuration, group int) []BenchListRow {
	result := make([]BenchListRow, len(benchlist))
	for i, bench := range benchlist {
		row := makeBenchResults(tree, bench, ignoredSpecs)
		row.Group = group
		result[i] = row
	}
	return result
}

func makeBenchResults(tree *TimeTree, bench benchbase.Benchmark, ignoredSpecs benchbase.Configuration) BenchListRow {
	var result BenchListRow

	result.Name = makeBenchName(bench.Conf, ignoredSpecs)
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

func makeBenchName(conf benchbase.Configuration, ignoredSpecs benchbase.Configuration) string {
	var result []string

	if conf["Host"] != ignoredSpecs["Host"] {
		result = append(result, conf["Host"])
	}

	if conf["Rev"] != ignoredSpecs["Rev"] {
		r, err := strconv.ParseInt(conf["Rev"], 10, 64)
		if err == nil {
			result = append(result, fmt.Sprintf("r%v", r))
		}
	}

	if conf["Threads"] != ignoredSpecs["Threads"] {
		result = append(result, "("+conf["Threads"]+" threads)")
	}

	return strings.Join(result, " ")
}
