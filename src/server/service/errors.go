package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata      = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken         = status.Errorf(codes.Unauthenticated, "invalid token")
	errCouldNotStartProcess = status.Error(codes.Internal, "could not start the process")
	errIDNotFound           = status.Error(codes.NotFound, "id was not found")
)
