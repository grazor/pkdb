package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/grazor/pkdb/pkg/provider"
	"gopkg.in/yaml.v2"
)

func (entry fsEntry) metaAbsolutePath() (string, bool) {
	var path string

	path = filepath.Join(entry.absolutePath(), ".yml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join(entry.absolutePath(), ".yaml")
	}

	exists := true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exists = false
	}

	return path, exists
}

func (entry fsEntry) Meta() (map[string]interface{}, error) {
	meta := make(map[string]interface{})
	path, ok := entry.metaAbsolutePath()
	if !ok {
		return meta, nil
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return meta, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("failed to read meta file %s", err),
		}
	}

	err = yaml.Unmarshal(contents, &meta)
	if err != nil {
		return meta, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("failed to parse yml %s", err),
		}
	}

	return meta, nil
}

func (entry fsEntry) SetMeta(map[string]interface{}) error {
	return nil
}
