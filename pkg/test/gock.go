package test

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	gogh "github.com/google/go-github/github"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	"math/rand"
	"strings"
	"testing"
)

const gitHubNotFound = `{
  "message": "Not Found",
  "documentation_url": "https://developer.github.com/v3/...."
}`

func MockGHCalls(t *testing.T, prjPath, branch string, files, langs SliceOfStrings, modifiers ...GockModifier) {
	var entries []gogh.TreeEntry
	for _, file := range files() {
		entries = append(entries, gogh.TreeEntry{
			SHA:  sha(file),
			Path: String(file),
		})
	}
	tree := gogh.Tree{
		SHA:       sha(files()...),
		Truncated: Boolean(false),
		Entries:   entries,
	}

	bytes, err := json.Marshal(tree)
	require.NoError(t, err)

	treeMock := gock.New("https://api.github.com").
		Get(fmt.Sprintf("/repos/%s/git/trees/%s", prjPath, branch))
	for _, modify := range modifiers {
		modify(treeMock)
	}
	treeMock.Reply(200).
		BodyString(string(bytes))

	languages := map[string]int{}
	for _, lang := range langs() {
		languages[lang] = rand.Int()
	}

	bytes, err = json.Marshal(languages)
	require.NoError(t, err)

	langMock := gock.New("https://api.github.com").
		Get(fmt.Sprintf("/repos/%s/languages", prjPath))
	for _, modify := range modifiers {
		modify(langMock)
	}
	langMock.Reply(200).
		BodyString(string(bytes))
}

func MockNotFoundGitHub(repoIdentifier string) {
	gock.New("https://api.github.com").
		Get(fmt.Sprintf("/repos/%s/.*", repoIdentifier)).
		Times(2).
		Reply(404).
		BodyString(gitHubNotFound)
}

type GockModifier func(mock *gock.Request)

func sha(files ...string) *string {
	return String(string(sha1.New().Sum([]byte(strings.Join(files, "-")))))
}

func String(value string) *string {
	return &value
}
func Boolean(value bool) *bool {
	return &value
}
