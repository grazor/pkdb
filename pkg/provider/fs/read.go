package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grazor/pkdb/pkg/provider"
)

func (entry fsEntry) absolutePath() string {
	return filepath.Join(entry.provider.basePath, entry.relativePath)
}

func (entry fsEntry) ID() string {
	//TODO: read id from attrs
	_, name := filepath.Split(entry.relativePath)
	return name
}

func (entry fsEntry) Name() string {
	_, name := filepath.Split(entry.relativePath)
	return name
}

func (entry fsEntry) Path() string {
	return entry.relativePath
}

func (entry fsEntry) Attrs() map[string]interface{} {
	//TODO: read and cache attrs
	return nil
}

func (entry fsEntry) Size() int64 {
	return entry.fileInfo.Size()
}

func (entry fsEntry) Time() time.Time {
	return entry.fileInfo.ModTime()
}

func (entry fsEntry) HasChildren() bool {
	return entry.fileInfo.IsDir()
}

func (entry fsEntry) Children() ([]provider.Entry, error) {
	if !entry.HasChildren() {
		return nil, nil
	}

	dir, err := os.Open(entry.absolutePath())
	if err != nil {
		return nil, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open %v", entry.absolutePath()),
		}
	}
	defer dir.Close()

	dirContents, err := dir.Readdir(-1)
	if err != nil {
		return nil, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to read dir %v", entry.absolutePath()),
		}
	}

	children := make([]provider.Entry, 0, len(dirContents))
	for _, childInfo := range dirContents {
		name := childInfo.Name()
		ext := strings.ToLower(filepath.Ext(name))
		rest := strings.TrimSuffix(name, ext)
		if ext == ".yml" || ext == ".yaml" {
			if rest == "" {
				// This file contains metadata for current dir
				continue
			}
			restPath := filepath.Join(entry.absolutePath(), rest)
			if _, err := os.Stat(restPath); err == nil {
				// This file contains metadata for file named `rest`
				continue
			}
		}

		child := fsEntry{
			provider:     entry.provider,
			relativePath: filepath.Join(entry.relativePath, name),
			fileInfo:     childInfo,
		}
		children = append(children, child)
	}
	return children, nil
}

func (entry fsEntry) Reader(off int64) (io.ReadCloser, error) {
	file, err := os.Open(entry.absolutePath())
	if err != nil {
		return nil, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open for reading %v", entry.absolutePath()),
		}
	}

	file.Seek(off, 0)
	if err != nil {
		return nil, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to seek %v", entry.absolutePath()),
		}
	}

	return file, nil
}
