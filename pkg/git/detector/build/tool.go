package build

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"regexp"
)

type Tool struct {
	Language        string
	Name            string
	ExpectedRegexps []*regexp.Regexp
	ExpectedFiles   []string
}

func NewDetectedBuildTool(language string, name string, detectedFiles []string) *v1alpha1.DetectedBuildType {
	return &v1alpha1.DetectedBuildType{
		Language:      language,
		Name:          name,
		DetectedFiles: detectedFiles,
	}
}

var Tools = []Tool{Maven, Gradle, Golang, Ruby, NodeJS, PHP, Python, Perl, Dotnet}

var Maven = Tool{
	Name:            "Maven",
	Language:        "java",
	ExpectedRegexps: regexps(`pom\.xml`),
	ExpectedFiles:   []string{"pom.xml"},
}

var Gradle = Tool{
	Name:            "Gradle",
	Language:        "java",
	ExpectedRegexps: regexps(`.*gradle.*`),
	ExpectedFiles:   []string{"build.gradle", "gradlew", "gradlew.bat"},
}

var Golang = Tool{
	Name:            "Golang",
	Language:        "go",
	ExpectedRegexps: regexps(`main\.go`, `Gopkg\.toml`, `glide\.yaml`),
	ExpectedFiles:   []string{"main.go", "Gopkg.toml", "glide.yaml"},
}

var Ruby = Tool{
	Name:            "Ruby",
	Language:        "ruby",
	ExpectedRegexps: regexps(`Gemfile`, `Rakefile`, `config\.ru`),
	ExpectedFiles:   []string{"Gemfile", "Rakefile", "config.ru"},
}

var NodeJS = Tool{
	Name:            "NodeJS",
	Language:        "javascript",
	ExpectedRegexps: regexps(`app\.json`, `package\.json`, `gulpfile\.js`, `Gruntfile\.js`),
	ExpectedFiles:   []string{"app.json", "package.json", "gulpfile.js", "Gruntfile.js"},
}

var PHP = Tool{
	Name:            "PHP",
	Language:        "php",
	ExpectedRegexps: regexps(`index\.php`, `composer\.json`),
	ExpectedFiles:   []string{"index.php", "composer.json"},
}

var Python = Tool{
	Name:            "Python",
	Language:        "python",
	ExpectedRegexps: regexps(`requirements\.txt`, `setup\.py`),
	ExpectedFiles:   []string{"requirements.txt", "setup.py"},
}

var Perl = Tool{
	Name:            "Perl",
	Language:        "perl",
	ExpectedRegexps: regexps(`index\.pl`, `cpanfile`),
	ExpectedFiles:   []string{"index.pl", "cpanfile"},
}

var Dotnet = Tool{
	Name:            "Dotnet",
	Language:        "C#",
	ExpectedRegexps: regexps(`project\.json`, `.*\.csproj`),
	ExpectedFiles:   []string{"project.json", "app.csproj"},
}

func regexps(values ...string) []*regexp.Regexp {
	var regexps []*regexp.Regexp
	for _, value := range values {
		regexps = append(regexps, regexp.MustCompile(value))
	}
	return regexps
}
