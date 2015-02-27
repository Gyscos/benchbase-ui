package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Gyscos/benchbase"
)

type BenchCompareTable struct {
	Category string
	Width    int

	Titles      [][]BenchListTableTitle
	BenchGroups [][]BenchListRow
}

func MakeCompareTables(host string, spec string, values string, ignore string, filter string, focus string, depth int) ([]BenchCompareTable, error) {
	resp, err := http.Get(fmt.Sprintf("%v/compare?spec=%v&values=%v&ignore=%v&filter=%v", host, spec, values, ignore, filter))
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

	if depth != 0 {
		for _, table := range data.Result {
			for _, group := range table {
				truncateDeepResults(group, depth)
			}
		}
	}

	result := make([]BenchCompareTable, len(data.Result))

	for i, table := range data.Result {
		result[i] = makeCompareTable(table, spec)
	}

	return result, nil
}

func makeCompareTable(groups [][]benchbase.Benchmark, spec string) BenchCompareTable {
	var result BenchCompareTable

	if len(groups) == 0 || len(groups[0]) == 0 {
		return result
	}

	tree := makeTimeTree(groups[0][0].Result)
	mergeSingleChilds(tree)
	computeDepthWidth(tree)

	result.Titles = makeTimeLabels(tree, groups[0][0].Conf, spec)
	result.Width = tree.width + 1

	return result
}
