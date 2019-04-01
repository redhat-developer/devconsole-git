package detector

import "regexp"

type BuildEnvStats struct {
	SortedLanguages    []string
	DetectedBuildTools []*DetectedBuildTool
}

type DetectedBuildTool struct {
	language      string
	name          string
	detectedFiles []string
}

type BuildTool struct {
	Language      string
	Name          string
	ExpectedFiles []*regexp.Regexp
}

func NewDetectedBuildTool(language string, name string, detectedFiles []string) *DetectedBuildTool {
	return &DetectedBuildTool{
		language:      language,
		name:          name,
		detectedFiles: detectedFiles,
	}
}

var BuildTools = []BuildTool{Maven, Gradle, Golang, Ruby, NodeJS, PHP, Python, Perl, Dotnet}

var Maven = BuildTool{
	Name:          "Maven",
	Language:      "java",
	ExpectedFiles: regexps(`pom\.xml`),
}

var Gradle = BuildTool{
	Name:          "Gradle",
	Language:      "java",
	ExpectedFiles: regexps(`.*gradle.*`),
}

var Golang = BuildTool{
	Name:          "Golang",
	Language:      "go",
	ExpectedFiles: regexps(`main\.go`, `Gopkg\.toml`, `glide\.yaml`),
}

var Ruby = BuildTool{
	Name:          "Ruby",
	Language:      "ruby",
	ExpectedFiles: regexps(`Gemfile`, `Rakefile`, `config\.ru`),
}

var NodeJS = BuildTool{
	Name:          "NodeJS",
	Language:      "javascript",
	ExpectedFiles: regexps(`app\.json`, `package\.json`, `gulpfile\.js`, `Gruntfile\.js`),
}

var PHP = BuildTool{
	Name:          "PHP",
	Language:      "php",
	ExpectedFiles: regexps(`index\.php`, `composer\.json`),
}

var Python = BuildTool{
	Name:          "Python",
	Language:      "python",
	ExpectedFiles: regexps(`requirements\.txt`, `setup\.py`),
}

var Perl = BuildTool{
	Name:          "Perl",
	Language:      "perl",
	ExpectedFiles: regexps(`index\.pl`, `cpanfile`),
}

var Dotnet = BuildTool{
	Name:          "Dotnet",
	Language:      "C#",
	ExpectedFiles: regexps(`project\.json`, `.*\.csproj`),
}

func regexps(values ...string) []*regexp.Regexp {
	var regexps []*regexp.Regexp
	for _, value := range values {
		regexps = append(regexps, regexp.MustCompile(value))
	}
	return regexps
}
