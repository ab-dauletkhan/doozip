package entities

type ArchiveInfo struct {
	Filename    string       `json:"filename"`
	ArchiveSize float64      `json:"archive_size"`
	TotalSize   float64      `json:"total_size"`
	TotalFiles  int          `json:"total_files"` // in the requirements it says float, but it should be int
	Files       []FileEntity `json:"files"`
}

type FileEntity struct {
	FilePath string  `json:"file_path"`
	Size     float64 `json:"size"`
	MimeType string  `json:"mimetype"`
}
