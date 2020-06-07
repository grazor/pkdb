package runner

import (
	"fmt"
	"log"
)

type RunnerError struct {
	Inner   error
	Message string
	Source  string
}

func (err RunnerError) Error() string {
	if err.Source != "" {
		return fmt.Sprintf("%v: %v", err.Source, err.Message)
	}
	return err.Message
}

func handleError(err error, message string) {
	log.Printf("%#v", err)
	fmt.Printf("%v", message)
}
