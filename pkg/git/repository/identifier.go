package repository

import (
	"errors"
	"fmt"
	"github.com/redhat-developer/git-service/pkg/apis/devconsole/v1alpha1"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
	"strings"
)

const Master = "master"

type StructuredIdentifier struct {
	Owner  string
	Name   string
	Branch string
}

func NewStructuredIdentifier(gitSource *v1alpha1.GitSource, endpoint *gittransport.Endpoint) (StructuredIdentifier, error) {
	var repo StructuredIdentifier
	branch := Master

	if gitSource.Spec.Ref != "" {
		branch = gitSource.Spec.Ref
	}
	path := endpoint.Path
	if strings.HasSuffix(path, ".git") {
		path = path[:len(path)-4]
	}

	switch urlSegments := strings.Split(path, "/"); len(urlSegments) {
	case 0, 1:
		return repo, errors.New("url is invalid")
	case 2:
		if urlSegments[0] == "" {
			return repo, errors.New("url is invalid")
		}
		return StructuredIdentifier{
			Owner:  urlSegments[0],
			Name:   urlSegments[1],
			Branch: branch,
		}, nil
	default:
		return StructuredIdentifier{
			Owner:  urlSegments[1],
			Name:   urlSegments[2],
			Branch: branch,
		}, nil
	}
}

func (i StructuredIdentifier) OwnerWithName() string {
	return fmt.Sprintf("%s/%s", i.Owner, i.Name)
}
