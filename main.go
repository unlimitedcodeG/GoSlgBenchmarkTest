package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/protobuf/proto"

	"GoSlgBenchmarkTest/internal/testserver"
	"GoSlgBenchmarkTest/internal/wsclient"
	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

func main() {
	var (
		mode     = flag.String("mode", "demo", "运行模式: demo, server, client")
		addr     = flag.String("addr", ":8080", "服务器地址")
		url      = flag.String("url", "ws://localhost:8080/ws", "WebSocket连接URL")
		token    = flag.String("token", "demo-token", "认证令牌")
		clients  = flag.Int("clients", 1, "客户端数量")
		duration = flag.Duration("duration", 30*time.Second, "运行时长")
	)
	flag.Parse()

	switch *mode {
	case "demo":
		runDemo()
	case "server":
		runServer(*addr)
	case "client":
		runClient(*url, *token, *clients, *duration)
	default:
		fmt.Printf("未知模式: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

// runDemo 运行演示模式
func runDemo() {
	fmt.Println("🚀 GoSlgBenchmarkTest - Unity长连接+Protobuf测试框架")
	fmt.Println("=================================================")
	fmt.Println()

	fmt.Println("📋 项目特性:")
	fmt.Println("  ✅ WebSocket长连接 + 自动重连")
	fmt.Println("  ✅ Protobuf消息序列化")
	fmt.Println("  ✅ 心跳机制 + RTT统计")
	fmt.Println("  ✅ 消息序列号去重")
	fmt.Println("  ✅ 完整测试套件(端到端/模糊/基准)")
	fmt.Println("  ✅ CI/CD Pipeline")
	fmt.Println()

	fmt.Println("🔧 快速开始:")
	fmt.Println("  # 生成Protobuf代码")
	fmt.Println("  make proto")
	fmt.Println()
	fmt.Println("  # 运行所有测试")
	fmt.Println("  make test")
	fmt.Println()
	fmt.Println("  # 运行基准测试")
	fmt.Println("  make bench")
	fmt.Println()
	fmt.Println("  # 启动测试服务器")
	fmt.Println("  go run main.go -mode=server")
	fmt.Println()
	fmt.Println("  # 运行客户端压力测试")
	fmt.Println("  go run main.go -mode=client -clients=10 -duration=60s")
	fmt.Println()

	fmt.Println("📚 更多信息:")
	fmt.Println("  make help    # 查看所有可用命令")
	fmt.Println("  make verify  # 完整项目验证")
}

// runServer 运行测试服务器
func runServer(addr string) {
	fmt.Printf("🖥️  启动测试服务器 %s\n", addr)

	config := testserver.DefaultServerConfig(addr)
	config.EnableBattlePush = true
	config.PushInterval = 100 * time.Millisecond

	server := testserver.New(config)

	if err := server.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	fmt.Printf("✅ 服务器已启动，监听地址: %s\n", addr)
	fmt.Printf("📊 统计信息: http://%s/stats\n", addr[1:]) // 去掉开头的冒号
	fmt.Printf("🎮 WebSocket端点: ws://%s/ws\n", addr[1:])

	// 优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Println("\n🔄 正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭错误: %v", err)
	}

	fmt.Println("✅ 服务器已关闭")
}

// runClient 运行客户端压力测试
func runClient(url, token string, clientCount int, duration time.Duration) {
	fmt.Printf("🔥 启动客户端压力测试\n")
	fmt.Printf("   连接URL: %s\n", url)
	fmt.Printf("   客户端数量: %d\n", clientCount)
	fmt.Printf("   运行时长: %v\n", duration)
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), duration+10*time.Second)
	defer cancel()

	// 统计信息
	stats := &ClientStats{}
	clients := make([]*wsclient.Client, clientCount)

	// 创建客户端
	for i := 0; i < clientCount; i++ {
		config := wsclient.DefaultClientConfig(url, fmt.Sprintf("%s-%d", token, i))
		config.HeartbeatInterval = 5 * time.Second

		client := wsclient.New(config)

		// 设置消息处理器
		client.SetPushHandler(func(opcode uint16, message proto.Message) {
			stats.AddMessage()
		})

		client.SetRTTHandler(func(rtt time.Duration) {
			stats.AddRTT(rtt)
		})

		client.SetStateChangeHandler(func(oldState, newState wsclient.ClientState) {
			if newState == wsclient.StateConnected {
				stats.AddConnection()
			} else if oldState == wsclient.StateConnected {
				stats.RemoveConnection()
			}
		})

		clients[i] = client
	}

	// 连接所有客户端
	fmt.Printf("🔗 正在连接 %d 个客户端...\n", clientCount)
	for i, client := range clients {
		if err := client.Connect(ctx); err != nil {
			log.Printf("客户端 %d 连接失败: %v", i, err)
		} else {
			fmt.Printf("✅ 客户端 %d 已连接\n", i)
		}
		time.Sleep(10 * time.Millisecond) // 避免连接风暴
	}

	// 运行压力测试
	fmt.Printf("\n🚀 开始压力测试，运行 %v...\n", duration)
	startTime := time.Now()

	// 定期发送操作
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		actionSeq := uint64(0)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for i, client := range clients {
					if time.Since(startTime) >= duration {
						return
					}

					actionSeq++
					action := &gamev1.PlayerAction{
						ActionSeq:       actionSeq,
						PlayerId:        fmt.Sprintf("player-%d", i),
						ActionType:      gamev1.ActionType_ACTION_TYPE_MOVE,
						ClientTimestamp: time.Now().UnixMilli(),
						ActionData: &gamev1.ActionData{
							Data: &gamev1.ActionData_Move{
								Move: &gamev1.MoveAction{
									TargetPosition: &gamev1.Position{
										X: float32((actionSeq % 100)),
										Y: float32((actionSeq * 2 % 100)),
										Z: 0,
									},
									MoveSpeed: 5.0,
								},
							},
						},
					}

					client.SendAction(action)
					stats.AddSentMessage()
				}
			}
		}
	}()

	// 定期打印统计
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				if elapsed >= duration {
					return
				}

				connections := stats.GetConnections()
				received := stats.GetReceivedMessages()
				sent := stats.GetSentMessages()
				avgRTT := stats.GetAverageRTT()

				fmt.Printf("📊 [%.0fs] 连接: %d, 已收: %d, 已发: %d, 平均RTT: %.1fms\n",
					elapsed.Seconds(), connections, received, sent, avgRTT.Seconds()*1000)
			}
		}
	}()

	// 等待测试完成
	time.Sleep(duration)

	// 最终统计
	fmt.Printf("\n📋 压力测试完成!\n")
	fmt.Printf("   运行时长: %v\n", duration)
	fmt.Printf("   活跃连接: %d/%d\n", stats.GetConnections(), clientCount)
	fmt.Printf("   接收消息: %d\n", stats.GetReceivedMessages())
	fmt.Printf("   发送消息: %d\n", stats.GetSentMessages())
	fmt.Printf("   平均RTT: %.1fms\n", stats.GetAverageRTT().Seconds()*1000)

	if received := stats.GetReceivedMessages(); received > 0 {
		throughput := float64(received) / duration.Seconds()
		fmt.Printf("   吞吐量: %.1f 消息/秒\n", throughput)
	}

	// 关闭所有客户端
	fmt.Printf("\n🔄 正在关闭客户端...\n")
	for i, client := range clients {
		if err := client.Close(); err != nil {
			log.Printf("客户端 %d 关闭错误: %v", i, err)
		}
	}

	fmt.Println("✅ 压力测试完成!")
}

// ClientStats 客户端统计信息
type ClientStats struct {
	connections      int
	receivedMessages int64
	sentMessages     int64
	rttSum           time.Duration
	rttCount         int64
	mu               sync.RWMutex
}

func (s *ClientStats) AddConnection() {
	s.mu.Lock()
	s.connections++
	s.mu.Unlock()
}

func (s *ClientStats) RemoveConnection() {
	s.mu.Lock()
	if s.connections > 0 {
		s.connections--
	}
	s.mu.Unlock()
}

func (s *ClientStats) AddMessage() {
	s.mu.Lock()
	s.receivedMessages++
	s.mu.Unlock()
}

func (s *ClientStats) AddSentMessage() {
	s.mu.Lock()
	s.sentMessages++
	s.mu.Unlock()
}

func (s *ClientStats) AddRTT(rtt time.Duration) {
	s.mu.Lock()
	s.rttSum += rtt
	s.rttCount++
	s.mu.Unlock()
}

func (s *ClientStats) GetConnections() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connections
}

func (s *ClientStats) GetReceivedMessages() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.receivedMessages
}

func (s *ClientStats) GetSentMessages() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sentMessages
}

func (s *ClientStats) GetAverageRTT() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.rttCount == 0 {
		return 0
	}
	return s.rttSum / time.Duration(s.rttCount)
}
