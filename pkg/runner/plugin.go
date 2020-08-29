package runner

import (
	"fmt"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/kdbplugin"
)

func Plugin(name string) (kdbplugin.KdbPlugin, error) {
	return nil, nil
}

func ConfigurePlugins(tree *kdb.KdbTree) error {
	plugins := tree.Provider.Plugins()
	for _, pluginName := range plugins {
		p, err := Plugin(pluginName)
		if err != nil {
			return RunnerError{
				Inner:   err,
				Message: fmt.Sprintf("failed to create plugin %s", pluginName),
			}
		}

		err = p.Init()
		if err != nil {
			return RunnerError{
				Inner:   err,
				Message: fmt.Sprintf("failed to initialize plugin %s", pluginName),
			}
		}
		tree.RegisterPlugin(pluginName, p)
	}
	return nil
}
