package test

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	"gopkg.in/h2non/gock.v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	host                 = "https://github.com"
	urlPath              = "/owner/repo/blob/master/"
	headerCheckerBaseURL = host + urlPath
)

func NewDummyServiceCreator(flavor string, shouldFail bool, files, langs SliceOfStrings) repository.ServiceCreator {
	return NewDummyService(flavor, shouldFail, files, langs, true).Creator()
}

func NewDummyServices(flavor string, shouldFail bool, files, langs SliceOfStrings) []*DummyService {
	return []*DummyService{
		NewDummyService(flavor, shouldFail, files, langs, false),
		NewDummyService(flavor, shouldFail, files, langs, true)}
}

func NewDummyService(flavor string, shouldFail bool, files, langs SliceOfStrings, useFilesChecker bool) *DummyService {
	return &DummyService{
		Files:           files(),
		Langs:           langs(),
		Flavor:          flavor,
		shouldFail:      shouldFail,
		UseFilesChecker: useFilesChecker,
	}
}

type DummyService struct {
	Files, Langs    []string
	shouldFail      bool
	Flavor          string
	UseFilesChecker bool
}

func (s *DummyService) FileExistenceChecker() (repository.FileExistenceChecker, error) {
	if s.Files == nil {
		return nil, fmt.Errorf("failing files")
	}
	if !s.UseFilesChecker {
		for _, file := range s.Files {
			gock.New(host).
				Head(urlPath + file + "$").
				Reply(200)
		}
		return repository.NewCheckerUsingHeaderRequests(&log.GitSourceLogger{Logger: logf.Log}, headerCheckerBaseURL,
			git.NewUsernamePassword("anynomous", "")), nil
	} else {
		return repository.NewCheckerWithFetchedFiles(s.Files), nil
	}
}
func (s *DummyService) GetLanguageList() ([]string, error) {
	if s.Langs == nil {
		return nil, fmt.Errorf("failing languages")
	}
	return s.Langs, nil
}

func (s *DummyService) Creator() repository.ServiceCreator {
	return func(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, secretProvider *git.SecretProvider) (service repository.GitService, e error) {
		if s.shouldFail {
			return nil, fmt.Errorf("creation failed")
		}
		if gitSource.Spec.Flavor == s.Flavor {
			return s, nil
		}
		return nil, nil
	}
}

func (s *DummyService) CheckCredentials() error {
	return nil
}
func (s *DummyService) CheckRepoAccessibility() error {
	return nil
}
func (s *DummyService) CheckBranch() error {
	return nil
}
