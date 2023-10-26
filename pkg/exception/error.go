package exception

import (
	"errors"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/logger"
)

type ErrorExceptions struct {
	Message string
	Err     error
}

func ErrorHandler(e *ErrorExceptions) error {
	if e != nil {
		logger.Log.Error(e.Err, e)
	}

	err := errors.New(e.Message)
	return err
}
