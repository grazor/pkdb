package serve

import (
	"log"

	"github.com/grazor/pkdb/pkg/load"
)

func Serve(source string, destinations []string) {
	//log.Print("Starting loader")
	loader, err := load.GetLoader(source)
	if err != nil {
		log.Fatal(err)
	}

	err = loader.Load(-1)
	if err != nil {
		log.Fatal(err)
	}

	//log.Print(loader.Children())
}
