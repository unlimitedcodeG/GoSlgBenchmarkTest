// Package main implements a standalone server for both gRPC and WebSocket testing
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"GoSlgBenchmarkTest/internal/grpcserver"
	"GoSlgBenchmarkTest/internal/testserver"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

func main() {
	fmt.Println("🎮 Unity SLG Test Server v2.0 (gRPC + WebSocket)")
	fmt.Println("=================================================")

	// 获取服务器类型
	serverType := os.Getenv("SERVER_TYPE")
	if serverType == "" {
		serverType = "websocket" // 默认启动WebSocket服务器用于CI测试
	}

	// 如果没有指定类型且有命令行参数，解析参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "grpc":
			serverType = "grpc"
		case "websocket", "ws":
			serverType = "websocket"
		case "help", "-h", "--help":
			fmt.Println("用法: go run tools/grpc-server.go [grpc|websocket]")
			fmt.Println("")
			fmt.Println("参数:")
			fmt.Println("  grpc      - 启动gRPC服务器 (端口19001)")
			fmt.Println("  websocket - 启动WebSocket服务器 (端口18090)")
			fmt.Println("  (无参数)  - 启动WebSocket服务器 (默认)")
			fmt.Println("")
			fmt.Println("环境变量:")
			fmt.Println("  SERVER_TYPE    - 服务器类型 (grpc|websocket)")
			fmt.Println("  GRPC_PORT      - gRPC服务器端口 (默认19001)")
			fmt.Println("  WS_PORT        - WebSocket服务器端口 (默认18090)")
			fmt.Println("  CI             - CI环境标识 (true时启用额外日志)")
			os.Exit(0)
		}
	}

	switch serverType {
	case "grpc":
		startGRPCServer()
	case "websocket":
		startWebSocketServer()
	default:
		fmt.Printf("❌ 未知的服务器类型: %s\n", serverType)
		os.Exit(1)
	}
}

func startGRPCServer() {
	fmt.Println("🚀 启动gRPC服务器模式...")

	// 默认端口19001，可以通过环境变量覆盖
	port := "19001"
	if envPort := os.Getenv("GRPC_PORT"); envPort != "" {
		port = envPort
	}

	// 监听指定端口
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("❌ 监听端口 %s 失败: %v", port, err)
	}
	defer lis.Close()

	// 创建gRPC服务器
	s := grpc.NewServer()

	// 创建游戏服务器实例
	gameServer := grpcserver.NewGameServer()

	// 注册服务
	gamev1.RegisterGameServiceServer(s, gameServer)

	// 启用反射（可选，用于调试）
	reflection.Register(s)

	fmt.Printf("✅ gRPC服务器启动成功!\n")
	fmt.Printf("📍 监听地址: 0.0.0.0:%s\n", port)
	fmt.Printf("🔧 服务端点: localhost:%s\n", port)
	fmt.Println("\n🎯 支持的gRPC方法:")
	fmt.Println("  • Login - 用户登录")
	fmt.Println("  • Logout - 用户登出")
	fmt.Println("  • GetPlayerStatus - 获取玩家状态")
	fmt.Println("  • SendPlayerAction - 发送玩家操作")
	fmt.Println("  • JoinBattle - 加入战斗")
	fmt.Println("  • GetBattleStatus - 获取战斗状态")
	fmt.Println("  • StreamBattleUpdates - 流式战斗更新")
	fmt.Println("  • StreamPlayerEvents - 流式玩家事件")
	fmt.Println("  • BatchPlayerActions - 批量玩家操作")
	fmt.Println("\n🧪 测试命令:")
	fmt.Println("  go run tools/grpc-test-client.go")
	fmt.Println("\n⏹️  按 Ctrl+C 停止服务器")

	// 启动服务器
	go func() {
		fmt.Printf("\n🚀 服务器正在运行...\n")
		if err := s.Serve(lis); err != nil {
			log.Printf("❌ gRPC服务器错误: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 正在关闭gRPC服务器...")

	// 优雅关闭
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.GracefulStop()
	}()

	// 设置超时
	select {
	case <-done:
		fmt.Println("✅ gRPC服务器已优雅关闭")
	case <-time.After(5 * time.Second):
		fmt.Println("⚠️  关闭超时，强制停止")
		s.Stop()
	}

	fmt.Println("👋 服务器已停止")
}

func startWebSocketServer() {
	fmt.Println("🚀 启动WebSocket服务器模式...")

	// 默认端口18090，可以通过环境变量覆盖
	port := "18090"
	if envPort := os.Getenv("WS_PORT"); envPort != "" {
		port = envPort
	}

	// 创建WebSocket测试服务器
	config := testserver.DefaultServerConfig(":" + port)
	server := testserver.New(config)

	// 启动服务器
	fmt.Printf("✅ WebSocket服务器启动成功!\n")
	fmt.Printf("📍 监听地址: ws://localhost:%s/ws\n", port)
	fmt.Printf("🔧 测试端点: ws://localhost:%s/ws\n", port)

	// 在CI环境中，输出环境变量供后续步骤使用
	if os.Getenv("CI") == "true" {
		fmt.Printf("🔧 CI环境变量: WS_PORT=%s\n", port)
		fmt.Printf("🔧 健康检查端点: http://localhost:%s/health\n", port)
	}
	fmt.Println("\n🎯 WebSocket服务器特性:")
	fmt.Println("  • 支持二进制消息传输")
	fmt.Println("  • 自动心跳和保活")
	fmt.Println("  • 连接状态管理")
	fmt.Println("  • 消息序列号验证")
	fmt.Println("  • 实时推送测试")
	fmt.Println("\n🧪 WebSocket测试命令:")
	fmt.Println("  go test ./test/session -run TestTimelineAnalysisRealWallClock -v")
	fmt.Println("  go test ./test/session -run TestSessionRecordingAndReplay -v")
	fmt.Println("  go test ./test/session -run TestSessionAssertions -v")
	fmt.Println("\n⏹️  按 Ctrl+C 停止服务器")

	// 启动服务器
	go func() {
		fmt.Printf("\n🚀 WebSocket服务器正在运行...\n")
		if err := server.Start(); err != nil {
			log.Printf("❌ WebSocket服务器错误: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 正在关闭WebSocket服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("⚠️  关闭WebSocket服务器时出错: %v\n", err)
	} else {
		fmt.Println("✅ WebSocket服务器已优雅关闭")
	}

	fmt.Println("👋 服务器已停止")
}
