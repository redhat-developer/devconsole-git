package repository

import (
	"encoding/base64"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/detector/build"
	"net/http"
	"regexp"
	"sync"
)

type checkerWithFetchedFiles struct {
	filenames []string
}

func NewCheckerWithFetchedFiles(filenames []string) FileExistenceChecker {
	return &checkerWithFetchedFiles{
		filenames: filenames,
	}
}

func (c *checkerWithFetchedFiles) GetListOfFoundFiles() []string {
	return c.filenames
}

func (c *checkerWithFetchedFiles) DetectFiles(buildTool build.Tool) []string {
	detectedFiles := make(chan string, len(buildTool.ExpectedRegexps))
	var wg sync.WaitGroup
	wg.Add(len(buildTool.ExpectedRegexps))

	for _, file := range buildTool.ExpectedRegexps {
		go func(buildToolFile *regexp.Regexp) {
			defer wg.Done()
			if detectedFile := getFileIfExists(buildToolFile, c.filenames); detectedFile != "" {
				detectedFiles <- detectedFile
			}
		}(file)
	}

	wg.Wait()
	close(detectedFiles)
	var result []string
	for detectedFile := range detectedFiles {
		if detectedFile != "" {
			result = append(result, detectedFile)
		}
	}

	return result
}

func getFileIfExists(buildToolFile *regexp.Regexp, actualFiles []string) string {
	for _, file := range actualFiles {
		if buildToolFile.MatchString(file) {
			return file
		}
	}
	return ""
}

type checkerUsingHeaderRequests struct {
	baseURL string
	secret  git.Secret
}

func NewCheckerUsingHeaderRequests(baseURL string, secret git.Secret) FileExistenceChecker {
	return &checkerUsingHeaderRequests{
		baseURL: baseURL,
		secret:  secret,
	}
}

func (c *checkerUsingHeaderRequests) GetListOfFoundFiles() []string {
	return []string{}
}

func (c *checkerUsingHeaderRequests) DetectFiles(buildTool build.Tool) []string {
	detectedFiles := make(chan string, len(buildTool.ExpectedFiles))
	var wg sync.WaitGroup
	wg.Add(len(buildTool.ExpectedFiles))
	client := c.secret.Client()

	for _, file := range buildTool.ExpectedFiles {
		go func(buildToolFile string) {
			defer wg.Done()
			if c.fileExists(client, buildToolFile) {
				detectedFiles <- buildToolFile
			}
		}(file)
	}

	wg.Wait()
	close(detectedFiles)
	var result []string
	for detectedFile := range detectedFiles {
		if detectedFile != "" {
			result = append(result, detectedFile)
		}
	}

	return result
}

func (c *checkerUsingHeaderRequests) fileExists(client *http.Client, buildToolFile string) bool {
	url := c.baseURL + buildToolFile
	request, err := NewRequest(http.MethodHead, url, c.secret)
	if err != nil {
		return false
	}
	resp, err := client.Do(request)
	if err != nil {
		return false
	} else {
		resp.Body.Close()
		return resp.StatusCode == 200
	}
}

func NewRequest(method, url string, secret git.Secret) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	if secret.SecretType() == git.UsernamePasswordType {
		req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(secret.SecretContent())))
	}
	return req, nil
}
