package test

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func AssertContainsBuildTool(t *testing.T, detected []v1alpha1.DetectedBuildType, name, language string, files ...string) {
	var found bool
	for _, tool := range detected {
		if tool.Name == name {
			assert.Equal(t, language, tool.Language)
			assert.Len(t, tool.DetectedFiles, len(files))
			for _, file := range files {
				assert.Contains(t, tool.DetectedFiles, file)
			}
			found = true
			break
		}
	}
	if !found {
		assert.Fail(t, fmt.Sprintf("the list %v does not contain tool with name %s, language %s and files %v",
			detected, name, language, files))
	}
}
