package main

import (
	"bytes"
	"sort"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	jobs := []string{
		"a", "b", "c", "d",
		"aa", "bb", "cc", "dd",
		"aaa", "bbb", "ccc", "ddd",
	}

	expected := make([]string, len(jobs))
	copy(expected, jobs)
	sort.Strings(expected)

	var w bytes.Buffer
	args := []string{
		"prodcons", /*"-prods", "2", "-conss", "3",/*/
	}
	args = append(args, jobs...)
	if err := run(&w, args); err != nil {
		t.Errorf("should have run: %s", err)
		return
	}

	actual := strings.Split(strings.TrimRight(w.String(), "\n"), "\n")
	sort.Strings(actual)
	if len(actual) != len(expected) {
		t.Errorf("unexpected len of outputs: got %d, but expect %d", len(actual), len(expected))
		return
	}
	for i := range actual {
		if actual[i] != expected[i] {
			t.Errorf("unexpected %d-th output: got %s, but expect %s", i, actual[i], expected[i])
			return
		}
	}
}
