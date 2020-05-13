package serve

import (
	"log"

	"github.com/grazor/pkdb/pkg/load"
)

type Server interface {
	Serve(load.TreeNode) (chan interface{}, error)
}

func Serve(source string, destinations []string) {
	root, err := load.GetNode(source)
	if err != nil {
		log.Fatal(err)
	}

	err = root.Load(-1)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Watching")
	doneWatch, err := root.Watch()
	if err != nil {
		log.Fatal(err)
	}
	defer close(doneWatch)

	<-doneWatch
}
