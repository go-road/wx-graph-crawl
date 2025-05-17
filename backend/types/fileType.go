package types

type SelectFileResponse struct {
	FilePath  string   `json:"file_path"`
	ValidURLs []string `json:"valid_urls"`
}
