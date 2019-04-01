package bitbucket

type Pagination struct {
	Page int    `json:"page,omitempty"`
	Next string `json:"next,omitempty"`
}

type Src struct {
	Pagination
	Values []FileEntry `json:"values,omitempty"`
}

type FileEntry struct {
	Path string `json:"path,omitempty"`
}

type RepositoryLanguage struct {
	Language string `json:"language,omitempty"`
}

type ResponseError struct {
	Error Error `json:"error,omitempty"`
}

type Error struct {
	Message string `json:"message,omitempty"`
}
