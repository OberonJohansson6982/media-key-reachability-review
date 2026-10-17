package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type keyRecord struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type auditRow struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	Reachable []string `json:"reachable"`
}

// Pipeline stages are local audit labels; the scopes are supplied by the operator.
var stages = []string{"asset ingestion", "processing jobs", "creator delivery"}

func audit(keys []keyRecord, required map[string][]string) []auditRow {
	rows := make([]auditRow, 0, len(keys))
	for _, key := range keys {
		granted := make(map[string]bool, len(key.Scopes))
		for _, scope := range key.Scopes {
			granted[scope] = true
		}
		row := auditRow{ID: key.ID, Name: key.Name, Scopes: key.Scopes, Reachable: []string{}}
		for _, stage := range stages {
			needed := required[stage]
			if len(needed) == 0 {
				continue
			}
			allowed := true
			for _, scope := range needed {
				if !granted[scope] {
					allowed = false
					break
				}
			}
			if allowed {
				row.Reachable = append(row.Reachable, stage)
			}
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	return rows
}

func reportMarkdown(date time.Time, rows []auditRow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Media key access review - %s\n\n", date.UTC().Format("2006-01-02"))
	b.WriteString("| Key ID | Name | Scopes | Reachable stages |\n| --- | --- | --- | --- |\n")
	for _, row := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", cell(row.ID), cell(row.Name), cell(strings.Join(row.Scopes, ", ")), cell(strings.Join(row.Reachable, ", ")))
	}
	return b.String()
}

func cell(s string) string {
	return strings.NewReplacer("|", "\\|", "\n", " ", "\r", " ").Replace(s)
}
