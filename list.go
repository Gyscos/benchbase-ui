package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/Gyscos/benchbase"
)

type ListBenchTable struct {
	Category string
	Width    int

	Titles    [][]BenchTableTitle
	BenchList BenchTableRows
}

func MakeListRequestURL(host string, filters string, ordering string, max int) string {
	return fmt.Sprintf("%v/list?filters=%v&ordering=%v&max=%v", host, filters, ordering, max)
}

func MakeListTables(requestURL string, focus string, depth int) ([]ListBenchTable, error) {
	resp, err := http.Get(requestURL)
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

	if len(data.Result) == 0 {
		return nil, nil
	}

	if focus != "" {
		cutOutOfFocus(data.Result, focus)
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

	result := make([]ListBenchTable, len(groups))
	for i, g := range groups {
		result[i] = makeListTable(g)
		result[i].BenchList.TableID = i
	}

	return result, nil
}

func groupByConf(benchlist []benchbase.Benchmark) [][]benchbase.Benchmark {
	m := make(map[string][]benchbase.Benchmark)
	for _, b := range benchlist {
		m[b.Conf["ForceAnalyze"]] = append(m[b.Conf["ForceAnalyze"]], b)
	}

	// log.Println(m)

	var keys []string
	for k, v := range m {
		if len(v) == 0 || len(v[0].Result) == 0 {
			delete(m, k)
		} else {
			keys = append(keys, k)
		}
	}

	sort.StringSlice(keys).Sort()

	var result [][]benchbase.Benchmark
	for _, k := range keys {
		result = append(result, m[k])
	}
	return result
}

func makeListTable(benchlist []benchbase.Benchmark) ListBenchTable {
	var result ListBenchTable

	if len(benchlist) == 0 {
		return result
	}

	tree := makeTimeTree(benchlist[0].Result)
	mergeSingleChilds(tree)
	computeDepthWidth(tree)

	result.Category = categoryName(benchlist[0].Conf)
	result.Titles = makeTimeLabels(tree, benchlist[0].Conf, "Host")
	result.BenchList = makeBenchGroup(tree, benchlist, benchbase.Configuration{}, 0)
	result.Width = tree.width + 1

	return result
}

func categoryName(conf benchbase.Configuration) string {
	if conf["ForceAnalyze"] == "true" {
		return "Analyze API"
	} else {
		return "Direct API"
	}
}
