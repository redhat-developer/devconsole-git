package git

import (
	"sort"
)

// SortLanguagesWithInts sorts languages from the given map to a slice where the first one is with the highest number
func SortLanguagesWithInts(languages map[string]int) []string {
	var langsWithRatio []langWithRatio
	for lang, ratio := range languages {
		langsWithRatio = append(langsWithRatio, langWithRatio{ratio: float64(ratio), lang: lang})
	}
	return sortLangs(langsWithRatio)
}

// SortLanguagesWithFloats32 sorts languages from the given map to a slice where the first one is with the highest number
func SortLanguagesWithFloats32(languages map[string]float32) []string {
	var langsWithRatio []langWithRatio
	for lang, ratio := range languages {
		langsWithRatio = append(langsWithRatio, langWithRatio{ratio: float64(ratio), lang: lang})
	}
	return sortLangs(langsWithRatio)
}

func sortLangs(langsWithRatio []langWithRatio) []string {
	sort.Sort(byRatio(langsWithRatio))

	var sortedLangs []string
	for _, sortedLang := range langsWithRatio {
		sortedLangs = append(sortedLangs, sortedLang.lang)
	}
	return sortedLangs
}

type langWithRatio struct {
	ratio float64
	lang  string
}

type byRatio []langWithRatio

func (a byRatio) Len() int      { return len(a) }
func (a byRatio) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byRatio) Less(i, j int) bool {
	return a[i].ratio > a[j].ratio
}
