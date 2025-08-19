package protocol

// 操作码定义 - 用于识别不同类型的消息
const (
	// 认证相关
	OpLoginReq  uint16 = 1001
	OpLoginResp uint16 = 1002
	OpLogout    uint16 = 1003

	// 心跳相关
	OpHeartbeat     uint16 = 1100
	OpHeartbeatResp uint16 = 1101

	// 游戏逻辑 - 战斗相关
	OpBattlePush   uint16 = 2001
	OpPlayerAction uint16 = 2002
	OpActionResp   uint16 = 2003

	// 聊天相关
	OpChatMessage uint16 = 3001
	OpChatResp    uint16 = 3002

	// 错误响应
	OpError uint16 = 9999
)

// OpcodeToString 将操作码转换为可读字符串，用于调试和日志
func OpcodeToString(op uint16) string {
	switch op {
	case OpLoginReq:
		return "LOGIN_REQ"
	case OpLoginResp:
		return "LOGIN_RESP"
	case OpLogout:
		return "LOGOUT"
	case OpHeartbeat:
		return "HEARTBEAT"
	case OpHeartbeatResp:
		return "HEARTBEAT_RESP"
	case OpBattlePush:
		return "BATTLE_PUSH"
	case OpPlayerAction:
		return "PLAYER_ACTION"
	case OpActionResp:
		return "ACTION_RESP"
	case OpChatMessage:
		return "CHAT_MESSAGE"
	case OpChatResp:
		return "CHAT_RESP"
	case OpError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// IsValidOpcode 检查操作码是否有效
func IsValidOpcode(op uint16) bool {
	switch op {
	case OpLoginReq, OpLoginResp, OpLogout,
		OpHeartbeat, OpHeartbeatResp,
		OpBattlePush, OpPlayerAction, OpActionResp,
		OpChatMessage, OpChatResp,
		OpError:
		return true
	default:
		return false
	}
}

// IsRequestOpcode 判断是否为请求类型的操作码
func IsRequestOpcode(op uint16) bool {
	switch op {
	case OpLoginReq, OpLogout, OpHeartbeat, OpPlayerAction, OpChatMessage:
		return true
	default:
		return false
	}
}

// IsResponseOpcode 判断是否为响应类型的操作码
func IsResponseOpcode(op uint16) bool {
	switch op {
	case OpLoginResp, OpHeartbeatResp, OpActionResp, OpChatResp, OpError:
		return true
	default:
		return false
	}
}

// IsPushOpcode 判断是否为推送类型的操作码
func IsPushOpcode(op uint16) bool {
	switch op {
	case OpBattlePush:
		return true
	default:
		return false
	}
}
