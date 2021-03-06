package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Gyscos/benchbase"
)

type CompareBenchTable struct {
	Category string
	Width    int

	Titles      [][]BenchTableTitle
	BenchGroups []BenchTableRows
}

func MakeCompareRequestURL(host string, spec string, values string, ignore string, filters string) string {
	return fmt.Sprintf("%v/compare?spec=%v&values=%v&ignore=%v&filter=%v", host, spec, values, ignore, filters)
}

func MakeCompareTables(requestURL string, spec string, focus string, depth int) ([]CompareBenchTable, error) {
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Error  string
		Result [][][]benchbase.Benchmark
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&data)

	if data.Error != "" {
		return nil, errors.New(data.Error)
	}

	if focus != "" {
		for _, table := range data.Result {
			for _, group := range table {
				cutOutOfFocus(group, focus)
			}
		}
		if depth != 0 {
			depth += 1 + strings.Count(focus, ".")
		}
	}

	if depth < 0 {
		// Find best depth
		depth = findBestDepth(getFirstResult(data.Result))
	}

	if depth != 0 {
		for _, table := range data.Result {
			for _, group := range table {
				truncateDeepResults(group, depth)
			}
		}
	}

	result := make([]CompareBenchTable, len(data.Result))

	for i, table := range data.Result {
		result[i] = makeCompareTable(table, spec, i)
	}

	return result, nil
}

func getFirstResult(tables [][][]benchbase.Benchmark) benchbase.Result {
	for _, table := range tables {
		for _, group := range table {
			for _, bench := range group {
				return bench.Result
			}
		}
	}

	return benchbase.Result{}
}

func makeCompareTable(groups [][]benchbase.Benchmark, spec string, tableId int) CompareBenchTable {
	var result CompareBenchTable

	if len(groups) == 0 || len(groups[0]) == 0 {
		return result
	}

	tree := makeTimeTree(groups[0][0].Result)
	mergeSingleChilds(tree)
	computeDepthWidth(tree)

	commonSpecs := getCommonSpecs(groups)
	result.Category = describeConf(commonSpecs)
	result.Titles = makeTimeLabels(tree, groups[0][0].Conf, spec)
	result.BenchGroups = makeBenchGroups(tree, groups, commonSpecs, tableId)
	result.Width = tree.width + 1

	return result
}

func intersectConf(a, b benchbase.Configuration) benchbase.Configuration {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	result := benchbase.Configuration{}

	for k, v := range a {
		if b[k] == v {
			result[k] = v
		}
	}

	return result
}

func getCommonSpecs(groups [][]benchbase.Benchmark) benchbase.Configuration {
	var commonSpecs benchbase.Configuration
	// We want the intersection of common specs
	for _, group := range groups {
		for _, bench := range group {
			commonSpecs = intersectConf(commonSpecs, bench.Conf)
		}
	}
	return commonSpecs
}

func describeConf(conf benchbase.Configuration) string {
	var result []string

	if conf["ForceAnalyze"] == "true" {
		result = append(result, "[Analyze API]")
	} else {
		result = append(result, "[Direct API]")
	}

	if conf["Host"] != "" {
		result = append(result, fmt.Sprintf("Host:%v", conf["Host"]))
	}

	if conf["Rev"] != "" {
		result = append(result, fmt.Sprintf("r%v", conf["Rev"]))
	}

	if conf["Threads"] != "" {
		result = append(result, fmt.Sprintf("(%v threads)", conf["Threads"]))
	}

	return strings.Join(result, " ")
}

func makeBenchGroups(tree *TimeTree, groups [][]benchbase.Benchmark, commonSpecs benchbase.Configuration, tableId int) []BenchTableRows {
	rowGroups := make([]BenchTableRows, len(groups))
	for i, group := range groups {
		rowGroups[i] = makeBenchGroup(tree, group, commonSpecs, i+1)
		rowGroups[i].TableID = i
	}
	return rowGroups
}
