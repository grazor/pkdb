package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

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

func (entry fsEntry) Delete() error {
	var err error

	if metaPath, ok := entry.metaAbsolutePath(); ok {
		err = syscall.Unlink(metaPath)
		if err != nil {
			return provider.ProviderError{
				Inner:   err,
				Message: fmt.Sprintf("unable to delete metadata for %v", entry.absolutePath()),
			}
		}
	}

	if entry.fileInfo.IsDir() {
		err = syscall.Rmdir(entry.absolutePath())
	} else {
		err = syscall.Unlink(entry.absolutePath())
	}

	if err != nil {
		return provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to delete %v", entry.absolutePath()),
		}
	}

	return nil
}

func (entry fsEntry) Move(targetParent provider.Entry, name string) error {
	target, ok := targetParent.(fsEntry)
	if !ok {
		return provider.ProviderError{
			Message: "move target is not a fsEntry",
		}
	}

	targetName := filepath.Join(target.absolutePath(), name)
	err := os.Rename(entry.absolutePath(), targetName)
	if err != nil {
		return provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("failed to move %s to %s", entry.absolutePath(), targetName),
		}
	}

	return nil
}
