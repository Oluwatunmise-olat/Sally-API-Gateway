package exception

import (
	"errors"
	"log"
)

type ErrorExceptions struct {
	Message string
	Err     error
}

func ErrorHandler(e *ErrorExceptions) error {
	if e != nil {
		log.Println(e.Err)
	}

	err := errors.New(e.Message)
	return err
}
