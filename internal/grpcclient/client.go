package grpcclient

import (
	"context"
	"fmt"
	"time"

	"HYH-Blog-Gin/proto/imageconvpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewImageClient 生成一个用于 ImageService 的 gRPC 客户端。
// 返回 imageconvpb.ImageServiceClient、cleanup 函数和错误。
func NewImageClient(ctx context.Context, addr string) (imageconvpb.ImageServiceClient, func(), error) {
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial imageconv gRPC server: %w", err)
	}

	client := imageconvpb.NewImageServiceClient(conn)
	cleanup := func() {
		_ = conn.Close()
	}
	return client, cleanup, nil
}
