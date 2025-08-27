package grpcserver

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gamev1 "GoSlgBenchmarkTest/proto/game/v1"
)

// GameServer gRPC游戏服务器实现
type GameServer struct {
	gamev1.UnimplementedGameServiceServer

	// 模拟数据存储
	players  sync.Map // map[string]*PlayerData
	battles  sync.Map // map[string]*BattleData
	sessions sync.Map // map[string]*SessionData

	// 统计信息
	requestCount int64
	startTime    time.Time
	mu           sync.RWMutex
}

type PlayerData struct {
	PlayerID   string
	Nickname   string
	Level      int32
	Experience int32
	Status     string
	Position   *gamev1.Position
	LastSeen   time.Time
}

type BattleData struct {
	BattleID    string
	Status      string
	Units       []*gamev1.BattleUnit
	StartTime   time.Time
	EndTime     *time.Time
	Subscribers map[string]chan *gamev1.BattlePush
	mu          sync.RWMutex
}

type SessionData struct {
	SessionID    string
	PlayerID     string
	LoginTime    time.Time
	LastActivity time.Time
}

// NewGameServer 创建新的游戏服务器实例
func NewGameServer() *GameServer {
	return &GameServer{
		startTime: time.Now(),
	}
}

// Login 用户登录
func (s *GameServer) Login(ctx context.Context, req *gamev1.LoginReq) (*gamev1.LoginResp, error) {
	// 简单的token验证
	if req.Token == "" {
		return nil, status.Error(codes.Unauthenticated, "token is required")
	}

	// 模拟登录处理时间 - 减少延迟以提高性能
	time.Sleep(2 * time.Millisecond)

	playerID := fmt.Sprintf("player_%s", req.Token)
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())

	// 存储会话信息
	s.sessions.Store(sessionID, &SessionData{
		SessionID:    sessionID,
		PlayerID:     playerID,
		LoginTime:    time.Now(),
		LastActivity: time.Now(),
	})

	// 存储或更新玩家信息
	s.players.Store(playerID, &PlayerData{
		PlayerID:   playerID,
		Nickname:   fmt.Sprintf("Player_%s", req.Token),
		Level:      1,
		Experience: 0,
		Status:     "online",
		Position:   &gamev1.Position{X: 0, Y: 0, Z: 0},
		LastSeen:   time.Now(),
	})

	return &gamev1.LoginResp{
		Ok:         true,
		PlayerId:   playerID,
		SessionId:  sessionID,
		ServerTime: time.Now().UnixMilli(),
	}, nil
}

// Logout 用户登出
func (s *GameServer) Logout(ctx context.Context, req *gamev1.LogoutReq) (*gamev1.LogoutResp, error) {
	if session, ok := s.sessions.Load(req.SessionId); ok {
		sessionData := session.(*SessionData)
		// 更新玩家状态
		if player, exists := s.players.Load(sessionData.PlayerID); exists {
			playerData := player.(*PlayerData)
			playerData.Status = "offline"
			playerData.LastSeen = time.Now()
			s.players.Store(sessionData.PlayerID, playerData)
		}
		s.sessions.Delete(req.SessionId)
	}

	return &gamev1.LogoutResp{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

// RefreshToken 刷新令牌
func (s *GameServer) RefreshToken(ctx context.Context, req *gamev1.RefreshTokenReq) (*gamev1.RefreshTokenResp, error) {
	// 模拟token验证和刷新
	if req.RefreshToken == "" {
		return nil, status.Error(codes.Unauthenticated, "refresh token is required")
	}

	newAccessToken := fmt.Sprintf("access_%d", time.Now().UnixNano())
	newRefreshToken := fmt.Sprintf("refresh_%d", time.Now().UnixNano())

	return &gamev1.RefreshTokenResp{
		Success:      true,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(time.Hour).UnixMilli(),
	}, nil
}

// SendPlayerAction 发送玩家操作
func (s *GameServer) SendPlayerAction(ctx context.Context, req *gamev1.PlayerAction) (*gamev1.ActionResp, error) {
	// 验证玩家是否存在
	if _, exists := s.players.Load(req.PlayerId); !exists {
		return nil, status.Error(codes.NotFound, "player not found")
	}

	// 模拟操作处理时间
	processingTime := time.Duration(5+req.ActionSeq%20) * time.Millisecond
	time.Sleep(processingTime)

	// 模拟一定的失败率（5%）
	if req.ActionSeq%20 == 0 {
		return &gamev1.ActionResp{
			Success:         false,
			ActionId:        req.ActionSeq,
			ErrorMessage:    "Action failed due to server error",
			ServerTimestamp: time.Now().UnixMilli(),
		}, nil
	}

	return &gamev1.ActionResp{
		Success:         true,
		ActionId:        req.ActionSeq,
		ServerTimestamp: time.Now().UnixMilli(),
	}, nil
}

// GetPlayerStatus 获取玩家状态
func (s *GameServer) GetPlayerStatus(ctx context.Context, req *gamev1.PlayerStatusReq) (*gamev1.PlayerStatusResp, error) {
	player, exists := s.players.Load(req.PlayerId)
	if !exists {
		return nil, status.Error(codes.NotFound, "player not found")
	}

	playerData := player.(*PlayerData)

	// 模拟库存物品
	inventory := []*gamev1.PlayerItem{
		{ItemId: "sword_001", ItemName: "Iron Sword", Quantity: 1, Rarity: "common"},
		{ItemId: "potion_001", ItemName: "Health Potion", Quantity: 5, Rarity: "common"},
		{ItemId: "armor_001", ItemName: "Leather Armor", Quantity: 1, Rarity: "uncommon"},
	}

	return &gamev1.PlayerStatusResp{
		PlayerId:        playerData.PlayerID,
		Status:          playerData.Status,
		CurrentPosition: playerData.Position,
		Level:           playerData.Level,
		Experience:      playerData.Experience,
		Inventory:       inventory,
	}, nil
}

// UpdatePlayerProfile 更新玩家资料
func (s *GameServer) UpdatePlayerProfile(ctx context.Context, req *gamev1.UpdateProfileReq) (*gamev1.UpdateProfileResp, error) {
	player, exists := s.players.Load(req.PlayerId)
	if !exists {
		return nil, status.Error(codes.NotFound, "player not found")
	}

	playerData := player.(*PlayerData)
	playerData.Nickname = req.Nickname
	s.players.Store(req.PlayerId, playerData)

	return &gamev1.UpdateProfileResp{
		Success:   true,
		Message:   "Profile updated successfully",
		UpdatedAt: time.Now().UnixMilli(),
	}, nil
}

// GetBattleStatus 获取战斗状态
func (s *GameServer) GetBattleStatus(ctx context.Context, req *gamev1.BattleStatusReq) (*gamev1.BattleStatusResp, error) {
	battle, exists := s.battles.Load(req.BattleId)
	if !exists {
		// 创建一个新的模拟战斗
		battleData := &BattleData{
			BattleID:    req.BattleId,
			Status:      "active",
			StartTime:   time.Now().Add(-time.Minute * 2),
			Subscribers: make(map[string]chan *gamev1.BattlePush),
		}

		// 生成模拟战斗单位
		if req.IncludeUnits {
			battleData.Units = []*gamev1.BattleUnit{
				{
					UnitId:   "unit_001",
					Hp:       85,
					Mp:       60,
					Position: &gamev1.Position{X: 10.5, Y: 20.3, Z: 0.0},
					Status:   gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
				},
				{
					UnitId:   "unit_002",
					Hp:       92,
					Mp:       30,
					Position: &gamev1.Position{X: 15.2, Y: 18.7, Z: 1.0},
					Status:   gamev1.UnitStatus_UNIT_STATUS_MOVING,
				},
			}
		}

		s.battles.Store(req.BattleId, battleData)
		battle = battleData
	}

	battleData := battle.(*BattleData)

	result := &gamev1.BattleResult{
		WinnerTeam: "team_red",
		PlayerScores: map[string]int32{
			"player_001": 1250,
			"player_002": 980,
		},
		TotalDurationMs: int32(time.Since(battleData.StartTime).Milliseconds()),
	}

	return &gamev1.BattleStatusResp{
		BattleId:  req.BattleId,
		Status:    battleData.Status,
		Units:     battleData.Units,
		StartTime: battleData.StartTime.UnixMilli(),
		Result:    result,
	}, nil
}

// JoinBattle 加入战斗
func (s *GameServer) JoinBattle(ctx context.Context, req *gamev1.JoinBattleReq) (*gamev1.JoinBattleResp, error) {
	// 验证玩家是否存在
	if _, exists := s.players.Load(req.PlayerId); !exists {
		return nil, status.Error(codes.NotFound, "player not found")
	}

	// 生成战斗ID
	battleID := fmt.Sprintf("battle_%s_%d", req.BattleType, time.Now().UnixNano())

	// 模拟匹配等待时间
	time.Sleep(50 * time.Millisecond)

	return &gamev1.JoinBattleResp{
		Success:            true,
		BattleId:           battleID,
		TeamAssigned:       "team_red",
		EstimatedStartTime: time.Now().Add(time.Second * 30).UnixMilli(),
	}, nil
}

// LeaveBattle 离开战斗
func (s *GameServer) LeaveBattle(ctx context.Context, req *gamev1.LeaveBattleReq) (*gamev1.LeaveBattleResp, error) {
	// 清理战斗订阅
	if battle, exists := s.battles.Load(req.BattleId); exists {
		battleData := battle.(*BattleData)
		battleData.mu.Lock()
		if ch, ok := battleData.Subscribers[req.PlayerId]; ok {
			close(ch)
			delete(battleData.Subscribers, req.PlayerId)
		}
		battleData.mu.Unlock()
	}

	return &gamev1.LeaveBattleResp{
		Success:        true,
		Message:        "Left battle successfully",
		PenaltyApplied: req.Reason == "quit", // 主动退出会有惩罚
	}, nil
}

// StreamBattleUpdates 流式战斗更新
func (s *GameServer) StreamBattleUpdates(req *gamev1.BattleStreamReq, stream gamev1.GameService_StreamBattleUpdatesServer) error {
	// 获取或创建战斗
	battle, exists := s.battles.Load(req.BattleId)
	if !exists {
		return status.Error(codes.NotFound, "battle not found")
	}

	battleData := battle.(*BattleData)

	// 创建订阅通道
	updateCh := make(chan *gamev1.BattlePush, 10)
	battleData.mu.Lock()
	battleData.Subscribers[req.PlayerId] = updateCh
	battleData.mu.Unlock()

	// 清理订阅
	defer func() {
		battleData.mu.Lock()
		if ch, ok := battleData.Subscribers[req.PlayerId]; ok {
			close(ch)
			delete(battleData.Subscribers, req.PlayerId)
		}
		battleData.mu.Unlock()
	}()

	// 发送初始状态
	initialPush := &gamev1.BattlePush{
		Seq:       1,
		BattleId:  req.BattleId,
		StateHash: []byte{1, 2, 3, 4},
		Units:     battleData.Units,
		Timestamp: time.Now().UnixMilli(),
	}

	if err := stream.Send(initialPush); err != nil {
		return err
	}

	// 模拟定期更新
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	seq := uint64(2)
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			// 生成模拟更新
			push := &gamev1.BattlePush{
				Seq:       seq,
				BattleId:  req.BattleId,
				StateHash: []byte{byte(seq), byte(seq + 1), byte(seq + 2), byte(seq + 3)},
				Units:     s.generateRandomUnits(),
				Timestamp: time.Now().UnixMilli(),
			}

			if err := stream.Send(push); err != nil {
				log.Printf("Failed to send battle update: %v", err)
				return err
			}
			seq++

			// 模拟战斗结束
			if seq > 20 {
				return nil
			}
		}
	}
}

// StreamPlayerEvents 流式玩家事件
func (s *GameServer) StreamPlayerEvents(req *gamev1.PlayerEventStreamReq, stream gamev1.GameService_StreamPlayerEventsServer) error {
	// 模拟玩家事件推送
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	eventTypes := []string{"level_up", "item_received", "achievement", "friend_request"}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			eventType := eventTypes[time.Now().Unix()%int64(len(eventTypes))]

			event := &gamev1.PlayerEventPush{
				PlayerId:  req.PlayerId,
				EventType: eventType,
				EventData: fmt.Sprintf(`{"type":"%s","value":%d}`, eventType, time.Now().Unix()),
				Timestamp: time.Now().UnixMilli(),
			}

			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}

// BatchPlayerActions 批量玩家操作
func (s *GameServer) BatchPlayerActions(ctx context.Context, req *gamev1.BatchActionReq) (*gamev1.BatchActionResp, error) {
	if len(req.Actions) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no actions provided")
	}

	results := make([]*gamev1.ActionResp, 0, len(req.Actions))
	successCount := 0

	for _, action := range req.Actions {
		// 重用单个操作的逻辑
		result, err := s.SendPlayerAction(ctx, action)
		if err != nil {
			if req.AtomicExecution {
				return nil, err // 原子操作，任何失败都回滚
			}
			// 非原子操作，记录错误但继续
			result = &gamev1.ActionResp{
				Success:         false,
				ActionId:        action.ActionSeq,
				ErrorMessage:    err.Error(),
				ServerTimestamp: time.Now().UnixMilli(),
			}
		}

		if result.Success {
			successCount++
		}
		results = append(results, result)
	}

	return &gamev1.BatchActionResp{
		Success:        successCount == len(req.Actions),
		ActionResults:  results,
		ProcessedCount: int32(len(results)),
	}, nil
}

// GetBatchBattleStatus 批量获取战斗状态
func (s *GameServer) GetBatchBattleStatus(ctx context.Context, req *gamev1.BatchBattleStatusReq) (*gamev1.BatchBattleStatusResp, error) {
	battles := make([]*gamev1.BattleStatusResp, 0, len(req.BattleIds))
	notFound := make([]string, 0)

	for _, battleID := range req.BattleIds {
		statusReq := &gamev1.BattleStatusReq{
			BattleId:     battleID,
			IncludeUnits: req.IncludeUnits,
		}

		status, err := s.GetBattleStatus(ctx, statusReq)
		if err != nil {
			notFound = append(notFound, battleID)
			continue
		}

		battles = append(battles, status)
	}

	return &gamev1.BatchBattleStatusResp{
		Battles:     battles,
		TotalCount:  int32(len(battles)),
		NotFoundIds: notFound,
	}, nil
}

// generateRandomUnits 生成随机战斗单位（用于模拟）
func (s *GameServer) generateRandomUnits() []*gamev1.BattleUnit {
	units := make([]*gamev1.BattleUnit, 2)
	for i := 0; i < 2; i++ {
		units[i] = &gamev1.BattleUnit{
			UnitId: fmt.Sprintf("unit_%03d", i+1),
			Hp:     int32(80 + (time.Now().UnixNano() % 20)),
			Mp:     int32(50 + (time.Now().UnixNano() % 30)),
			Position: &gamev1.Position{
				X: float32(10 + (time.Now().UnixNano() % 20)),
				Y: float32(15 + (time.Now().UnixNano() % 20)),
				Z: 0,
			},
			Status: gamev1.UnitStatus_UNIT_STATUS_ATTACKING,
		}
	}
	return units
}

// GetStats 获取服务器统计信息
func (s *GameServer) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	playerCount := 0
	s.players.Range(func(key, value interface{}) bool {
		playerCount++
		return true
	})

	battleCount := 0
	s.battles.Range(func(key, value interface{}) bool {
		battleCount++
		return true
	})

	sessionCount := 0
	s.sessions.Range(func(key, value interface{}) bool {
		sessionCount++
		return true
	})

	return map[string]interface{}{
		"uptime_seconds":  time.Since(s.startTime).Seconds(),
		"total_players":   playerCount,
		"active_battles":  battleCount,
		"active_sessions": sessionCount,
		"request_count":   s.requestCount,
	}
}
