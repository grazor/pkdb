package run

import (
	"fmt"
	"log"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/load"
	"github.com/grazor/pkdb/pkg/serve"
)

func getLoader(source string) (kdb.LoadableTreeNode, error) {
	root, err := load.New(source)
	if err != nil {
		return nil, err
	}

	loadableRoot, ok := root.(kdb.LoadableTreeNode)
	if !ok {
		return nil, fmt.Errorf("source %v does not support reading", source)
	}
	return loadableRoot, nil
}

func getServers(destinations []string) ([]serve.Server, []error) {
	servers, errors := make([]serve.Server, 0, len(destinations)), make([]error, 0)

	for _, destination := range destinations {
		server, err := serve.New(destination)
		if err != nil {
			errors = append(errors, err)
		} else {
			servers = append(servers, server)

		}
	}
	return servers, errors
}

func Run(source string, destinations []string) {
	root, err := getLoader(source)
	if err != nil {
		log.Fatal(err)
	}
	err = root.Load(-1)
	if err != nil {
		log.Fatal(err)
	}
	done, err := root.Watch()
	if err != nil {
		log.Fatal(err)
	}
	defer close(done)

	servers, errors := getServers(destinations)
	for _, err := range errors {
		log.Println(err)
	}
	if len(servers) == 0 {
		log.Fatal("No servers available")
	}
	for _, server := range servers {
		err = server.Serve(root, done)
		if err != nil {
			log.Println(err)
		}
	}

	<-done
}
