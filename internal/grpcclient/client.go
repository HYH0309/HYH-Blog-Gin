package grpcclient

import (
	"context"
	"fmt"
	"time"

	"HYH-Blog-Gin/proto/imageconvpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// NewImageClient 生成一个用于 ImageService 的 gRPC 客户端。
// 返回 imageconvpb.ImageServiceClient、cleanup 函数和错误。
func NewImageClient(ctx context.Context, addr string) (imageconvpb.ImageServiceClient, func(), error) {
	// 使用新 API：grpc.NewClient（不直接接受上下文）。
	// 我们用连接状态等待配合外部 ctx 来实现 3 秒内连通性检查。
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	// 在 3 秒内等待连接进入 Ready 状态，否则认为不可用
	waitCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	state := conn.GetState()
	for state != connectivity.Ready {
		// 等待状态变化或超时
		if !conn.WaitForStateChange(waitCtx, state) {
			// 超时或 ctx 取消
			_ = conn.Close()
			return nil, nil, fmt.Errorf("grpc connection to %s not ready within timeout", addr)
		}
		state = conn.GetState()
	}

	client := imageconvpb.NewImageServiceClient(conn)
	cleanup := func() { _ = conn.Close() }
	return client, cleanup, nil
}
