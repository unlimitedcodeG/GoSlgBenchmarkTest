package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"GoSlgBenchmarkTest/internal/config"
	"GoSlgBenchmarkTest/internal/session"
	"GoSlgBenchmarkTest/internal/wsclient"

	"google.golang.org/protobuf/proto"
)

// 命令行参数
var (
	envFlag      = flag.String("env", "", "指定环境类型 (development|testing|staging|local)")
	configFlag   = flag.String("config", "", "配置文件路径 (默认: configs/test-environments.yaml)")
	playerFlag   = flag.String("player", "", "指定玩家用户名 (使用配置文件中的测试账号)")
	durationFlag = flag.Duration("duration", 30*time.Minute, "录制时长")
	outputFlag   = flag.String("output", "", "输出目录 (默认使用配置文件中的设置)")
	verboseFlag  = flag.Bool("verbose", false, "启用详细日志")
	dryRunFlag   = flag.Bool("dry-run", false, "干运行模式，只显示配置不实际录制")
)

func main() {
	flag.Parse()

	fmt.Println("🎮 Unity游戏录制工具")
	fmt.Println("=====================")
	fmt.Println()

	// 加载配置文件
	testConfig, err := config.LoadConfig(*configFlag)
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	fmt.Printf("✅ 配置文件加载成功: %s v%s\n",
		testConfig.Meta.Project, testConfig.Meta.ConfigVersion)

	// 确定使用的环境
	var envType config.EnvironmentType
	if *envFlag != "" {
		envType = config.EnvironmentType(*envFlag)
		if !envType.IsValid() {
			log.Fatalf("❌ 无效的环境类型: %s\n可用环境: development, testing, staging, local", *envFlag)
		}
	} else {
		envType = testConfig.DefaultEnvironment
		fmt.Printf("🔧 使用默认环境: %s\n", envType)
	}

	// 获取环境配置
	env, err := testConfig.GetEnvironment(envType)
	if err != nil {
		log.Fatalf("❌ 获取环境配置失败: %v", err)
	}

	fmt.Printf("🌍 目标环境: %s (%s)\n", env.Name, env.Description)
	fmt.Printf("🔗 服务器地址: %s\n", env.Server.WsURL)

	// 选择测试账号
	var testAccount *config.TestAccount
	if *playerFlag != "" {
		testAccount, err = env.GetTestAccount(*playerFlag)
		if err != nil {
			log.Fatalf("❌ 获取测试账号失败: %v", err)
		}
	} else {
		testAccount, err = env.GetFirstTestAccount()
		if err != nil {
			log.Fatalf("❌ 没有可用的测试账号: %v", err)
		}
		fmt.Printf("🔧 使用默认测试账号: %s\n", testAccount.Username)
	}

	fmt.Printf("👤 测试账号: %s (PlayerID: %s)\n",
		testAccount.Username, testAccount.PlayerID)

	// 确定输出目录
	outputDir := *outputFlag
	if outputDir == "" {
		outputDir = testConfig.Global.Recording.OutputDir
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("❌ 创建输出目录失败: %v", err)
	}

	fmt.Printf("📁 输出目录: %s\n", outputDir)
	fmt.Printf("⏱️  录制时长: %v\n", *durationFlag)

	// 干运行模式
	if *dryRunFlag {
		fmt.Println("\n🔍 干运行模式 - 配置检查完成")
		fmt.Println("配置验证通过，可以开始实际录制")
		return
	}

	// 开始录制
	fmt.Println("\n🎬 开始录制...")
	if err := startRecording(testConfig, env, testAccount, outputDir, *durationFlag); err != nil {
		log.Fatalf("❌ 录制失败: %v", err)
	}

	fmt.Println("\n🎉 录制完成！")
}

// startRecording 开始录制会话
func startRecording(testConfig *config.TestEnvironmentConfig, env *config.Environment,
	testAccount *config.TestAccount, outputDir string, duration time.Duration) error {

	// 创建会话录制器
	sessionID := fmt.Sprintf("%s_%s_%d",
		env.Testing.SessionPrefix, testAccount.Username, time.Now().Unix())
	recorder := session.NewSessionRecorder(sessionID)

	fmt.Printf("📹 会话ID: %s\n", sessionID)

	// 配置WebSocket客户端
	clientConfig := &wsclient.ClientConfig{
		URL:               env.Server.WsURL,
		Token:             testAccount.Token,
		ClientVersion:     testConfig.UnityClient.ClientInfo.Version,
		DeviceID:          fmt.Sprintf("%s_%s", testConfig.UnityClient.ClientInfo.DeviceIDPrefix, testAccount.Username),
		HandshakeTimeout:  env.Network.HandshakeTimeout,
		HeartbeatInterval: env.Network.HeartbeatInterval,
		PingTimeout:       env.Network.PingTimeout,
		ReconnectInterval: env.Network.ReconnectInterval,
		MaxReconnectTries: env.Network.MaxReconnectTries,
		EnableCompression: env.Network.EnableCompression,
		UserAgent:         testConfig.UnityClient.ClientInfo.UserAgent,
	}

	client := wsclient.New(clientConfig)

	// 设置消息处理器
	if env.Testing.EnableRecording {
		client.SetPushHandler(func(opcode uint16, message proto.Message) {
			recorder.RecordEvent(session.EventMessageReceive, map[string]interface{}{
				"opcode":       opcode,
				"message_type": fmt.Sprintf("%T", message),
				"player_id":    testAccount.PlayerID,
				"environment":  string(env.Name),
			})

			if *verboseFlag {
				fmt.Printf("📥 接收消息: opcode=%d, type=%T\n", opcode, message)
			}
		})

		client.SetStateChangeHandler(func(oldState, newState wsclient.ClientState) {
			recorder.RecordEvent(session.EventConnect, map[string]interface{}{
				"old_state":   oldState.String(),
				"new_state":   newState.String(),
				"player_id":   testAccount.PlayerID,
				"environment": string(env.Name),
			})

			fmt.Printf("🔄 状态变化: %s -> %s\n", oldState, newState)
		})

		// 设置RTT监听器
		client.SetRTTHandler(func(rtt time.Duration) {
			recorder.RecordLatency(rtt)
			if *verboseFlag {
				fmt.Printf("📊 RTT: %v\n", rtt)
			}
		})
	}

	// 连接服务器
	ctx, cancel := context.WithTimeout(context.Background(), env.Network.HandshakeTimeout)
	defer cancel()

	fmt.Printf("🔗 连接服务器: %s\n", env.Server.WsURL)
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("连接服务器失败: %w", err)
	}
	defer client.Close()

	fmt.Printf("✅ 连接成功\n")

	// 记录登录事件
	recorder.RecordEvent(session.EventLogin, map[string]interface{}{
		"player_id":   testAccount.PlayerID,
		"username":    testAccount.Username,
		"environment": string(env.Name),
		"server_url":  env.Server.WsURL,
		"status":      "success",
	})

	// 设置定时器
	timer := time.NewTimer(duration)
	defer timer.Stop()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("⏳ 录制进行中... (时长: %v)\n", duration)
	fmt.Println("💡 按 Ctrl+C 提前结束录制")

	// 定期输出统计信息
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	// 等待录制结束
	for {
		select {
		case <-timer.C:
			fmt.Println("⏰ 录制时间到，正在结束...")
			goto cleanup

		case <-sigChan:
			fmt.Println("\n🛑 收到中断信号，正在结束录制...")
			goto cleanup

		case <-ticker.C:
			elapsed := time.Since(startTime)
			remaining := duration - elapsed
			if remaining > 0 {
				fmt.Printf("📊 录制进度: %.1f%% (剩余: %v)\n",
					float64(elapsed)/float64(duration)*100, remaining.Round(time.Second))
			}

			// 注意：客户端连接状态通过StateChangeHandler监控
		}
	}

cleanup:
	// 记录关闭事件
	recorder.RecordClose(session.CloseNormal, "Recording completed")

	// 获取录制的会话
	recordedSession := recorder.GetSession()
	fmt.Printf("📋 录制统计: %d 个事件, %d 个消息帧\n",
		len(recordedSession.Events), len(recordedSession.Frames))

	// 导出会话数据
	return exportSession(testConfig, recorder, sessionID, outputDir)
}

// exportSession 导出会话数据
func exportSession(testConfig *config.TestEnvironmentConfig, recorder *session.SessionRecorder,
	sessionID, outputDir string) error {

	// 导出JSON格式
	if contains(testConfig.Global.Recording.ExportFormat, "json") {
		jsonData, err := recorder.ExportJSON()
		if err != nil {
			return fmt.Errorf("导出JSON失败: %w", err)
		}

		jsonFile := filepath.Join(outputDir, fmt.Sprintf("session_%s.json", sessionID))
		if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
			return fmt.Errorf("保存JSON文件失败: %w", err)
		}

		fmt.Printf("💾 会话数据已保存: %s (%d 字节)\n", jsonFile, len(jsonData))
	}

	// 如果启用了自动断言，运行断言测试
	if testConfig.Global.Assertions.MessageOrder.Enabled ||
		testConfig.Global.Assertions.Latency.Enabled ||
		testConfig.Global.Assertions.Reconnect.Enabled ||
		testConfig.Global.Assertions.ErrorRate.Enabled {

		fmt.Println("🧪 运行自动断言测试...")
		runAssertions(testConfig, recorder.GetSession())
	}

	return nil
}

// runAssertions 运行断言测试
func runAssertions(testConfig *config.TestEnvironmentConfig, recordedSession *session.Session) {
	suite := session.NewAssertionSuite("Unity Recording Quality Test", "Unity录制质量自动测试")

	// 消息顺序断言
	if testConfig.Global.Assertions.MessageOrder.Enabled {
		// 这里需要根据实际的协议操作码来配置
		// 示例使用通用的操作码
		suite.AddAssertion(session.NewMessageOrderAssertion(
			"Message Order Check",
			"验证消息按顺序接收",
			2002, // 这里应该使用配置文件中定义的操作码
			testConfig.Global.Assertions.MessageOrder.MinMessages,
			testConfig.Global.Assertions.MessageOrder.MaxMessages,
		))
	}

	// 延迟断言
	if testConfig.Global.Assertions.Latency.Enabled {
		suite.AddAssertion(session.NewLatencyAssertion(
			"Latency Check",
			"验证延迟在可接受范围内",
			testConfig.Global.Assertions.Latency.MaxLatency,
			testConfig.Global.Assertions.Latency.Percentile,
		))
	}

	// 重连断言
	if testConfig.Global.Assertions.Reconnect.Enabled {
		suite.AddAssertion(session.NewReconnectAssertion(
			"Reconnect Check",
			"验证重连次数和耗时",
			testConfig.Global.Assertions.Reconnect.MaxCount,
			testConfig.Global.Assertions.Reconnect.MaxDuration,
		))
	}

	// 错误率断言
	if testConfig.Global.Assertions.ErrorRate.Enabled {
		suite.AddAssertion(session.NewErrorRateAssertion(
			"Error Rate Check",
			"验证错误率在可接受范围内",
			testConfig.Global.Assertions.ErrorRate.MaxRate,
		))
	}

	// 运行断言
	suite.RunAssertions(recordedSession)

	passedCount := suite.GetPassedCount()
	failedCount := suite.GetFailedCount()
	successRate := suite.GetSuccessRate()

	fmt.Printf("📊 断言结果: %d 通过, %d 失败, 成功率: %.1f%%\n",
		passedCount, failedCount, successRate*100)

	if failedCount > 0 {
		fmt.Println("⚠️  存在失败的断言，请检查录制质量")
	}
}

// contains 检查切片中是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
