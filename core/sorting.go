package core

import "sort"

type SortBy int64

const (
	SortByName SortBy = iota
	SortBySize
)

func SortEntries(sortBy SortBy, entries []*Entry) {
	sort.Slice(entries, func(i, j int) bool {
		switch sortBy {
		case SortBySize:
			return entries[i].Size > entries[j].Size
		case SortByName:
			fallthrough
		default:
			return entries[i].Name < entries[j].Name
		}
	})
}
