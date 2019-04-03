package test

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
)

func NewDummyServiceCreator(flavor string, shouldFail bool, files, langs SliceOfStrings) repository.ServiceCreator {
	return NewDummyService(flavor, shouldFail, files, langs).Creator()
}

func NewDummyService(flavor string, shouldFail bool, files, langs SliceOfStrings) *DummyService {
	return &DummyService{
		Files:      files(),
		Langs:      langs(),
		Flavor:     flavor,
		shouldFail: shouldFail,
	}
}

type DummyService struct {
	Files, Langs []string
	shouldFail   bool
	Flavor       string
}

func (s *DummyService) GetListOfFilesInRootDir() ([]string, error) {
	if s.Files == nil {
		return nil, fmt.Errorf("failing files")
	}
	return s.Files, nil
}
func (s *DummyService) GetLanguageList() ([]string, error) {
	if s.Langs == nil {
		return nil, fmt.Errorf("failing languages")
	}
	return s.Langs, nil
}

func (s *DummyService) Creator() repository.ServiceCreator {
	return func(gitSource *v1alpha1.GitSource, secret git.Secret) (service repository.GitService, e error) {
		if s.shouldFail {
			return nil, fmt.Errorf("creation failed")
		}
		if gitSource.Spec.Flavor == s.Flavor {
			return s, nil
		}
		return nil, nil
	}
}
