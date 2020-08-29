package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/grazor/pkdb/pkg/provider"
)

const (
	defaultCreateMode = 0660
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

func (prov *fsProvider) Get(relativePath string) (provider.Entry, error) {
	fullPath := filepath.Join(prov.basePath, relativePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to stat %v", fullPath),
		}
	}

	entry := fsEntry{
		provider:     prov,
		relativePath: relativePath,
		fileInfo:     info,
	}

	return entry, nil
}
