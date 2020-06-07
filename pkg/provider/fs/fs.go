package fs

import (
	"os"
	"path/filepath"

	"github.com/grazor/pkdb/pkg/provider"
)

type fsProvider struct {
	basePath string
}

type fsEntry struct {
	provider     *fsProvider
	relativePath string
	fileInfo     os.FileInfo
}

func New(path string) (*fsProvider, error) {
	provider := &fsProvider{
		basePath: path,
	}
	return provider, nil
}

func (provider *fsProvider) Get(relativePath string) (provider.Entry, error) {
	fullPath := filepath.Join(provider.basePath, relativePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	entry := fsEntry{
		provider:     provider,
		relativePath: relativePath,
		fileInfo:     info,
	}
	return entry, nil
}
