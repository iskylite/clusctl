package utils

import (
	"google.golang.org/grpc/status"
)

func GrpcErrorMsg(err error) string {
	return status.Code(err).String()
	// return status.Convert(err).Message()
}
