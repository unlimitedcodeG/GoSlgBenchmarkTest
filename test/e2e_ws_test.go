package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"GoSlgBenchmarkTest/internal/testutil"
	wsclient "GoSlgBenchmarkTest/internal/wsclient"
)

// TestBasicConnection 测试基本连接功能
func TestBasicConnection(t *testing.T) {
	// 创建服务器
	server := testutil.NewTestServer(t)
	server.Start()
	defer server.Stop()

	// 创建客户端配置
	config := &wsclient.ClientConfig{
		URL:               server.GetWebSocketURL(),
		Token:             "test-token",
		ClientVersion:     "1.0.0",
		DeviceID:          "test-device",
		HandshakeTimeout:  5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		PingTimeout:       5 * time.Second,
		ReconnectInterval: 1 * time.Second,
		MaxReconnectTries: 3,
		EnableCompression: true,
		UserAgent:         "TestClient/1.0",
	}

	// 创建客户端
	client := wsclient.New(config)
	defer client.Close()

	// 直接连接
	t.Logf("🔌 Connecting to: %s", config.URL)
	err := client.Connect(context.Background())
	require.NoError(t, err, "Connection should succeed")

	// 验证连接状态
	time.Sleep(100 * time.Millisecond)
	stats := client.GetStats()
	t.Logf("🔍 Connection state: %s", stats["state"])

	// 状态验证
	if stats["state"] != "CONNECTED" {
		t.Errorf("Expected connection state CONNECTED, got %s", stats["state"])
	} else {
		t.Logf("✅ Connection test passed!")
	}

	// 保持连接一段时间
	time.Sleep(2 * time.Second)
}
