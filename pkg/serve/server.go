package serve

import (
	"log"

	"github.com/grazor/pkdb/pkg/load"
)

func Serve(source string, destinations []string) {
	root, err := load.GetNode(source)
	if err != nil {
		log.Fatal(err)
	}

	err = root.Load(-1)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(root)
}
