package runner

import (
	"fmt"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/plugins/extensions/markdown"
)

func Plugin(name string) (kdb.KdbPlugin, error) {
	switch name {
	case "markdown":
		return markdown.New(), nil
	}
	return nil, RunnerError{Message: fmt.Sprintf("unknown plugin %s", name)}
}

func ConfigurePlugins(tree *kdb.KdbTree) error {
	plugins := tree.Provider.Plugins()
	fmt.Println("Got plugins:", plugins)
	for _, pluginName := range plugins {
		p, err := Plugin(pluginName)
		if err != nil {
			return RunnerError{
				Inner:   err,
				Message: fmt.Sprintf("failed to create plugin %s", pluginName),
			}
		}

		err = p.Init(tree)
		if err != nil {
			return RunnerError{
				Inner:   err,
				Message: fmt.Sprintf("failed to initialize plugin %s", pluginName),
			}
		}
		fmt.Println("Registering plugin", pluginName)
		tree.RegisterPlugin(pluginName, p)
	}
	return nil
}
