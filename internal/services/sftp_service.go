package services

import (
	"fmt"
	"sort"
	"time"

	"github.com/pkg/sftp"
)

// FileInfo frontend structure.
type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

// ListDirectory returns a list of files at the specified path.
func ListDirectory(sftpClient *sftp.Client, path string) ([]FileInfo, error) {
	if sftpClient == nil {
		return nil, fmt.Errorf("SFTP client is not initialized")
	}

	// Reading the contents of the folder.
	files, err := sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var result []FileInfo
	for _, f := range files {
		result = append(result, FileInfo{
			Name:    f.Name(),
			Size:    f.Size(),
			Mode:    f.Mode().String(),
			ModTime: f.ModTime(),
			IsDir:   f.IsDir(),
		})
	}

	// Sort: first folders, then files (alphabetically).
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return result[i].Name < result[j].Name
	})

	return result, nil
}
