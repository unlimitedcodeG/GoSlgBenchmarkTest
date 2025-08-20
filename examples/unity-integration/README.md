# Unity真实集成方案

## 🎯 概述

这个文档说明如何将录制回放框架与真实Unity客户端和游戏服务器集成。

## 🔧 当前状况

### ❌ 当前是模拟环境：
```text
Go模拟客户端 ⟷ Go模拟服务器
     ↓              ↓
   录制回放      模拟游戏逻辑
```

### ✅ 目标真实环境：
```text
Unity客户端 ⟷ 真实游戏服务器
     ↓              ↓
   录制工具      真实游戏逻辑
```

## 🚀 集成方案

### 方案1：Unity客户端 + 真实服务器 + 并行录制

```text
┌─────────────┐    WebSocket    ┌──────────────┐
│Unity客户端   │ ◄─────────────► │真实游戏服务器│
└─────────────┘                 └──────────────┘
       │                               │
       │                               │
    游戏操作                         游戏响应
       │                               │
       ▼                               ▼
┌─────────────┐    监听/录制     ┌──────────────┐
│录制工具     │ ◄─────────────► │网络监听器     │
└─────────────┘                 └──────────────┘
```

**实施步骤：**
1. Unity客户端连接真实游戏服务器
2. 录制工具通过网络抓包或API监听
3. 记录所有网络交互和游戏事件

### 方案2：代理模式录制

```text
┌─────────────┐    WebSocket    ┌──────────────┐    WebSocket    ┌──────────────┐
│Unity客户端   │ ◄─────────────► │录制代理服务器│ ◄─────────────► │真实游戏服务器│
└─────────────┘                 └──────────────┘                 └──────────────┘
                                        │
                                        ▼
                                   录制所有消息
```

**实施步骤：**
1. 录制代理服务器监听端口（如8080）
2. Unity客户端连接代理服务器
3. 代理服务器转发所有消息到真实服务器
4. 同时录制所有经过的消息

### 方案3：Unity插件集成

```text
┌─────────────────────────────┐
│      Unity客户端            │
│  ┌─────────────────────────┐│    WebSocket    ┌──────────────┐
│  │   游戏逻辑              ││ ◄─────────────► │真实游戏服务器│
│  └─────────────────────────┘│                 └──────────────┘
│  ┌─────────────────────────┐│
│  │   录制插件              ││
│  │   - 监听网络事件        ││
│  │   - 记录用户操作        ││
│  │   - 导出测试数据        ││
│  └─────────────────────────┘│
└─────────────────────────────┘
```

## 📋 具体实施建议

### 推荐方案2：代理模式

**原因：**
- ✅ 不需要修改Unity代码
- ✅ 不需要修改游戏服务器
- ✅ 完全透明的录制
- ✅ 可以录制所有网络交互

**实施步骤：**

#### 1. 创建录制代理服务器
```go
// cmd/recording-proxy/main.go
type RecordingProxy struct {
    listenAddr    string          // Unity连接的地址
    targetAddr    string          // 真实游戏服务器地址
    recorder      *session.SessionRecorder
}
```

#### 2. Unity客户端配置
```csharp
// Unity中修改连接地址
// 从: ws://real-game-server.com:9090/ws
// 改为: ws://localhost:8080/ws  (录制代理地址)
```

#### 3. 启动录制
```bash
# 启动录制代理
go run cmd/recording-proxy/main.go \
  --listen=:8080 \
  --target=ws://real-game-server.com:9090/ws \
  --session-id=unity_test_session

# Unity客户端连接 localhost:8080
# 所有消息自动录制
```

## 🎮 Unity集成示例代码

### C# 网络管理器示例
```csharp
public class NetworkManager : MonoBehaviour 
{
    [SerializeField] private bool enableRecording = false;
    [SerializeField] private string serverUrl = "ws://real-server.com:9090/ws";
    [SerializeField] private string recordingProxyUrl = "ws://localhost:8080/ws";
    
    void Start() 
    {
        string targetUrl = enableRecording ? recordingProxyUrl : serverUrl;
        ConnectToServer(targetUrl);
    }
    
    void ConnectToServer(string url) 
    {
        // WebSocket连接逻辑
        // ...
        
        if (enableRecording) 
        {
            Debug.Log($"[Recording] 连接录制代理: {url}");
            // 可选：通知录制工具开始录制
            StartExternalRecording();
        }
    }
    
    void StartExternalRecording() 
    {
        // 通过HTTP API通知录制工具
        var request = UnityWebRequest.Post("http://localhost:8081/start-recording", "");
        request.SendWebRequest();
    }
}
```

## 🔍 测试点分析

### 当前模拟环境的测试点：
1. ✅ **协议兼容性** - 验证protobuf序列化/反序列化
2. ✅ **网络稳定性** - 测试重连、心跳机制
3. ✅ **性能基准** - 延迟、吞吐量测试
4. ✅ **消息顺序** - 验证消息按序处理
5. ✅ **错误处理** - 测试异常情况恢复

### 真实环境额外的测试点：
1. 🎯 **真实游戏逻辑** - 实际业务流程验证
2. 🎯 **Unity客户端兼容性** - 真实客户端行为
3. 🎯 **服务器负载** - 真实服务器性能
4. 🎯 **网络环境影响** - 真实网络条件
5. 🎯 **用户体验** - 实际游戏体验测试

## 📊 价值对比

| 测试场景 | 模拟环境 | 真实环境 |
|---------|---------|---------|
| 协议测试 | ✅ 完全支持 | ✅ 完全支持 |
| 性能测试 | ✅ 基础性能 | ✅ 真实性能 |
| 稳定性测试 | ✅ 网络层面 | ✅ 端到端 |
| 业务逻辑测试 | ❌ 模拟逻辑 | ✅ 真实逻辑 |
| 用户体验测试 | ❌ 无法测试 | ✅ 完整测试 |
| 调试便利性 | ✅ 完全可控 | ⚠️ 需要配合 |

## 🚀 下一步行动建议

1. **短期**：继续使用模拟环境完善录制回放框架
2. **中期**：实现录制代理服务器
3. **长期**：与Unity团队配合集成真实环境

这样既能快速验证框架功能，又能为真实集成做好准备！