package git

type Source struct {
	// URL of the git repo
	URL string

	// Ref is a git reference. Optional. "Master" is used by default.
	Ref string

	// Secret refers to the credentials to access the git repo. Optional.
	Secret Secret

	// Flavor of the git provider like github, gitlab, bitbucket, generic, etc. Optional.
	Flavor string
}
