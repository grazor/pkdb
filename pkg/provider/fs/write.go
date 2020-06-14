package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/grazor/pkdb/pkg/provider"
)

func (entry fsEntry) Writer(off int64) (io.WriteCloser, error) {
	file, err := os.OpenFile(entry.absolutePath(), os.O_WRONLY|os.O_CREATE, 0751)
	if err != nil {
		return nil, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open for writing %v", entry.absolutePath()),
		}
	}

	_, err = file.Seek(off, 0)
	if err != nil {
		return nil, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to seek %v", entry.absolutePath()),
		}
	}

	return file, nil
}

func (entry fsEntry) AddChild(name string, container bool) (provider.Entry, error) {
	childPath := filepath.Join(entry.relativePath, name)
	newEntry := fsEntry{provider: entry.provider, relativePath: childPath}
	childAsbolutePath := newEntry.absolutePath()

	if container {
		err := os.Mkdir(childAsbolutePath, 0751)
		if err != nil {
			return newEntry, provider.ProviderError{
				Inner:   err,
				Message: fmt.Sprintf("failed to create dir %v", childAsbolutePath),
			}
		}
	} else {
		file, err := os.Create(childAsbolutePath)
		if err != nil {
			return newEntry, provider.ProviderError{
				Inner:   err,
				Message: fmt.Sprintf("failed to create file %v", childAsbolutePath),
			}
		}
		defer file.Close()
	}

	info, err := os.Stat(childAsbolutePath)
	if err != nil {
		return newEntry, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to stat %v", childAsbolutePath),
		}
	}

	newEntry.fileInfo = info
	return newEntry, nil
}

func (entry fsEntry) Move()   {}
func (entry fsEntry) Delete() {}
