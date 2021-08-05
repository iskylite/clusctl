package utils

import (
	"errors"

	"google.golang.org/grpc/status"
)

func GrpcErrorMsg(err error) string {
	return status.Convert(err).Message()
	// return status.Code(err).String()
}

func GrpcErrorWrapper(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(GrpcErrorMsg(err))
}
