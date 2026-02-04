package server

import (
	"context"
	"net"
	"strings"

	"go.uber.org/zap"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TrustedSubnetUnaryInterceptor(trusted *net.IPNet, log *zap.Logger) grpcLib.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpcLib.UnaryServerInfo, handler grpcLib.UnaryHandler) (any, error) {
		if trusted == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		vals := md.Get("x-real-ip")
		if len(vals) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}

		ip := net.ParseIP(strings.TrimSpace(vals[0]))
		if ip == nil || !trusted.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "untrusted ip")
		}

		return handler(ctx, req)
	}
}
