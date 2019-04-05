package git_test

import (
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSortLanguagesWithInts(t *testing.T) {
	// given
	langs := map[string]int{"Go": 1, "Java": 2, "Ruby": 3}

	// when
	languages := git.SortLanguagesWithInts(langs)

	// then
	require.Len(t, languages, 3)
	assert.Equal(t, "Ruby", languages[0])
	assert.Equal(t, "Java", languages[1])
	assert.Equal(t, "Go", languages[2])
}

func TestSortLanguagesWithBiggerInts(t *testing.T) {
	// given
	langs := map[string]int{"Go": 345, "Java": 2345, "Ruby": 12345}

	// when
	languages := git.SortLanguagesWithInts(langs)

	// then
	require.Len(t, languages, 3)
	assert.Equal(t, "Ruby", languages[0])
	assert.Equal(t, "Java", languages[1])
	assert.Equal(t, "Go", languages[2])
}

func TestSortLanguagesWithIntsWhenAllNumbersAreSame(t *testing.T) {
	// given
	langs := map[string]int{"Go": 1, "Java": 1, "Ruby": 1}

	// when
	languages := git.SortLanguagesWithInts(langs)

	// then
	require.Len(t, languages, 3)
	assert.Contains(t, languages, "Ruby")
	assert.Contains(t, languages, "Java")
	assert.Contains(t, languages, "Go")
}

func TestSortLanguagesWithIntsWhenOneNumberIsBigger(t *testing.T) {
	// given
	langs := map[string]int{"Go": 3, "Java": 1, "Ruby": 1}

	// when
	languages := git.SortLanguagesWithInts(langs)

	// then
	require.Len(t, languages, 3)
	assert.Equal(t, "Go", languages[0])
	assert.Contains(t, languages, "Ruby")
	assert.Contains(t, languages, "Java")
}

func TestSortLanguagesWithFloats32(t *testing.T) {
	// given
	langs := map[string]float32{"Go": 1.9, "Java": 2.0, "Ruby": 2.0001}

	// when
	languages := git.SortLanguagesWithFloats32(langs)

	// then
	require.Len(t, languages, 3)
	assert.Equal(t, "Ruby", languages[0])
	assert.Equal(t, "Java", languages[1])
	assert.Equal(t, "Go", languages[2])
}

func TestSortLanguagesWithBiggerFloats32(t *testing.T) {
	// given
	langs := map[string]float32{"Go": 3.45, "Java": 23.45, "Ruby": 123.45}

	// when
	languages := git.SortLanguagesWithFloats32(langs)

	// then
	require.Len(t, languages, 3)
	assert.Equal(t, "Ruby", languages[0])
	assert.Equal(t, "Java", languages[1])
	assert.Equal(t, "Go", languages[2])
}

func TestSortLanguagesWithFloats32WhenAllNumbersAreSame(t *testing.T) {
	// given
	langs := map[string]float32{"Go": 1.1, "Java": 1.1, "Ruby": 1.1}

	// when
	languages := git.SortLanguagesWithFloats32(langs)

	// then
	require.Len(t, languages, 3)
	assert.Contains(t, languages, "Ruby")
	assert.Contains(t, languages, "Java")
	assert.Contains(t, languages, "Go")
}

func TestSortLanguagesWithFloats32WhenOneNumberIsBigger(t *testing.T) {
	// given
	langs := map[string]float32{"Go": 3.1, "Java": 1.1, "Ruby": 1.1}

	// when
	languages := git.SortLanguagesWithFloats32(langs)

	// then
	require.Len(t, languages, 3)
	assert.Equal(t, "Go", languages[0])
	assert.Contains(t, languages, "Ruby")
	assert.Contains(t, languages, "Java")
}
