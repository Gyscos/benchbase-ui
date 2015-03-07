package main

import (
	"sort"
	"strings"

	"github.com/Gyscos/benchbase"
)

type TimeTree struct {
	// Prefix:
	prefix   string
	name     string
	depth    int
	width    int
	children []*TimeTree
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
func getLabels(times benchbase.Result) []string {
	var result []string
	for k, _ := range times {
		result = append(result, k)
	}

	sort.StringSlice(result).Sort()

	return result
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

func findChild(node TimeTree, name string) int {
	for i, child := range node.children {
		if child.name == name {
			return i
		}
	}
	return -1
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
