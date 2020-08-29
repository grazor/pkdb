package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/grazor/pkdb/pkg/provider"
	"gopkg.in/yaml.v2"
)

var _ provider.Provider = (*fsProvider)(nil)

const (
	defaultCreateMode = 0660
)

type fsProvider struct {
	basePath string
	config   map[string]interface{}
}

type fsEntry struct {
	provider     *fsProvider
	relativePath string
	fileInfo     os.FileInfo
}

func New(path string) (*fsProvider, error) {
	p := &fsProvider{
		basePath: path,
	}

	err := p.configure()
	if err != nil {
		return nil, provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to configure provider %s", err),
		}
	}

	return p, nil
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

// Plugins returns names of plugins listed in the config.
func (prov *fsProvider) Plugins() []string {
	p := prov.config["plugins"]

	switch data := p.(type) {
	case []interface{}:
		plugins := make([]string, 0, len(data))
		for _, p := range data {
			plugins = append(plugins, fmt.Sprint(p))
		}
		return plugins

	case map[string]interface{}:
		plugins := make([]string, 0, len(data))
		for k := range data {
			plugins = append(plugins, k)
		}
		return plugins
	}

	return nil
}

func (prov *fsProvider) configure() error {
	configPath := filepath.Join(prov.basePath, "config.yml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		prov.config = make(map[string]interface{})
		return nil
	} else if err != nil {
		return provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to load config %s", err),
		}
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to read config %s", err),
		}
	}

	err = yaml.Unmarshal(configData, &prov.config)
	if err != nil {
		return provider.ProviderError{
			Inner:   err,
			Message: fmt.Sprintf("unable to process config %s", err),
		}
	}

	return nil
}
