package git

import (
	"sort"
)

// SortLanguagesWithInts sorts languages from the given map to a slice where the first one is with the highest number
func SortLanguagesWithInts(langsWithSizes map[string]int) []string {
	var contentSizes []int
	reversedMap := make(map[int][]string)
	for lang, size := range langsWithSizes {
		if _, ok := reversedMap[size]; !ok {
			contentSizes = append(contentSizes, size)
		}
		reversedMap[size] = append(reversedMap[size], lang)
	}
	sort.Ints(contentSizes)

	var sortedLangs []string
	for i := len(contentSizes) - 1; i >= 0; i-- {
		sortedLangs = append(sortedLangs, reversedMap[contentSizes[i]]...)
	}
	return sortedLangs
}

// SortLanguagesWithFloats32 sorts languages from the given map to a slice where the first one is with the highest number
func SortLanguagesWithFloats32(langsWithSizes map[string]float32) []string {
	var contentSizes []float64
	reversedMap := make(map[float64][]string)
	for lang, size32 := range langsWithSizes {
		size := float64(size32)
		if _, ok := reversedMap[size]; !ok {
			contentSizes = append(contentSizes, size)
		}
		reversedMap[size] = append(reversedMap[size], lang)
	}
	sort.Float64s(contentSizes)

	var sortedLangs []string
	for i := len(contentSizes) - 1; i >= 0; i-- {
		sortedLangs = append(sortedLangs, reversedMap[contentSizes[i]]...)
	}
	return sortedLangs
}
