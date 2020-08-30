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

	absPath := entry.absolutePath()
	if entry.fileInfo.IsDir() {
		path = filepath.Join(absPath, ".yml")
	} else {
		path = absPath + ".yml"
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if entry.fileInfo.IsDir() {
			path = filepath.Join(absPath, ".yaml")
		} else {
			path = absPath + ".yaml"
		}
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

func (entry fsEntry) SetMeta(data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}

	ymlData, err := yaml.Marshal(data)
	if err != nil {
		return provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("Unable to encode yaml data %s", err),
		}
	}

	path, _ := entry.metaAbsolutePath()
	err = ioutil.WriteFile(path, ymlData, defaultCreateMode)
	if err != nil {
		return provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("Unable to write metadata %s", err),
		}
	}

	return nil
}

func (entry fsEntry) UpdateMeta(data map[string]interface{}) error {
	meta, err := entry.Meta()
	if err != nil {
		return err
	}

	for k, v := range data {
		meta[k] = v
	}
	return entry.SetMeta(meta)
}
