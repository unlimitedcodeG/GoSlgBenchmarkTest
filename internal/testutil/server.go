package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"GoSlgBenchmarkTest/internal/config"
	"GoSlgBenchmarkTest/internal/testserver"
)

// TestServer 测试服务器包装器
type TestServer struct {
	*testserver.Server
	config *config.TestConfig
	addr   string
	t      *testing.T
}

// NewTestServer 创建测试服务器
func NewTestServer(t *testing.T) *TestServer {
	cfg := config.GetTestConfig()

	addr, err := cfg.GetServerAddress()
	require.NoError(t, err, "Failed to allocate server port")

	serverConfig := testserver.DefaultServerConfig(addr)
	serverConfig.EnableBattlePush = cfg.Server.Push.EnableBattlePush
	serverConfig.PushInterval = cfg.Server.Push.Interval

	server := testserver.New(serverConfig)

	ts := &TestServer{
		Server: server,
		config: cfg,
		addr:   addr,
		t:      t,
	}

	// 确保每个测试实例记录分配的端口，便于调试
	t.Logf("🔌 Test server allocated address: %s", addr)

	return ts
}

// Start 启动测试服务器
func (ts *TestServer) Start() {
	err := ts.Server.Start()
	require.NoError(ts.t, err, "Failed to start test server")

	// 等待服务器就绪 - 增加等待时间确保完全启动
	time.Sleep(2 * time.Second) // 增加等待时间到2秒

	ts.t.Logf("✅ Test server started on %s", ts.addr)
}

// Stop 停止测试服务器
func (ts *TestServer) Stop() {
	if ts.Server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), ts.config.Server.DefaultTimeout)
		defer cancel()

		ts.Server.Shutdown(ctx)

		// 等待服务器完全释放端口
		time.Sleep(200 * time.Millisecond)

		ts.config.ReleaseServerPort(ts.addr)
		ts.t.Logf("🛑 Test server stopped and port released: %s", ts.addr)
	}
}

// GetAddress 获取服务器地址
func (ts *TestServer) GetAddress() string {
	return ts.addr
}

// GetWebSocketURL 获取WebSocket URL
func (ts *TestServer) GetWebSocketURL() string {
	return ts.config.GetWebSocketURL(ts.addr)
}

// GetHTTPURL 获取HTTP URL
func (ts *TestServer) GetHTTPURL() string {
	return fmt.Sprintf("http://%s", ts.addr)
}

// WithCustomConfig 使用自定义配置创建测试服务器
func NewTestServerWithConfig(t *testing.T, customizer func(*testserver.ServerConfig)) *TestServer {
	ts := NewTestServer(t)

	// 重新创建服务器配置
	serverConfig := testserver.DefaultServerConfig(ts.addr)
	customizer(serverConfig)

	ts.Server = testserver.New(serverConfig)
	return ts
}

// StartWithTimeout 带超时启动服务器
func (ts *TestServer) StartWithTimeout(timeout time.Duration) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- ts.Server.Start()
	}()

	select {
	case err := <-errCh:
		if err == nil {
			time.Sleep(ts.config.TestScenarios.BasicConnection.ValidationDelay)
			ts.t.Logf("✅ Test server started on %s", ts.addr)
		}
		return err
	case <-time.After(timeout):
		return fmt.Errorf("server start timeout after %v", timeout)
	}
}

// WaitForReady 等待服务器就绪
func (ts *TestServer) WaitForReady() {
	time.Sleep(ts.config.TestScenarios.BasicConnection.ValidationDelay)
}

// Cleanup 清理资源（用于测试完成后）
func (ts *TestServer) Cleanup() {
	if ts.Server != nil {
		ts.Stop()
	}
}
