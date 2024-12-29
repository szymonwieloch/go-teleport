package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func ensureValidToken(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errMissingMetadata
		}
		// The keys within metadata.MD are normalized to lowercase.
		// See: https://godoc.org/google.golang.org/grpc/metadata#New
		if !validSecret(md["authorization"], secret) {
			return nil, errInvalidToken
		}
		// Continue execution of handler after ensuring a valid token.
		return handler(ctx, req)
	}
}

func validSecret(authorization []string, secret string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	// Perform the token validation here. For the sake of this example, the code
	// here forgoes any of the usual OAuth2 token validation and instead checks
	// for a token matching an arbitrary string.
	return token == secret
}

func configOAuth(opts []grpc.ServerOption, args args) ([]grpc.ServerOption, error) {
	cert, err := tls.LoadX509KeyPair(args.AuthCert, args.AuthKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}
	return append(opts,
		grpc.UnaryInterceptor(ensureValidToken(args.Secret)),
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	), nil

}
