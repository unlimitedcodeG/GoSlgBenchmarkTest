package session

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ReplaySpeed 回放速度
type ReplaySpeed float64

const (
	SpeedSlow    ReplaySpeed = 0.5 // 慢速回放
	SpeedNormal  ReplaySpeed = 1.0 // 正常速度
	SpeedFast    ReplaySpeed = 2.0 // 快速回放
	SpeedInstant ReplaySpeed = 0.0 // 瞬间回放（无延迟）
)

// ReplayConfig 回放配置
type ReplayConfig struct {
	Speed         ReplaySpeed   `json:"speed"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	EnablePause   bool          `json:"enable_pause"`
	PauseOnError  bool          `json:"pause_on_error"`
	MaxReplayTime time.Duration `json:"max_replay_time"`
	EventFilter   EventFilter   `json:"event_filter"`
}

// EventFilter 事件过滤器
type EventFilter struct {
	EventTypes []EventType    `json:"event_types"`
	Opcode     *uint16        `json:"opcode,omitempty"`
	MinLatency *time.Duration `json:"min_latency,omitempty"`
	MaxLatency *time.Duration `json:"max_latency,omitempty"`
}

// ReplayEvent 回放事件
type ReplayEvent struct {
	OriginalEvent *SessionEvent `json:"original_event"`
	ReplayTime    time.Time     `json:"replay_time"`
	Delay         time.Duration `json:"delay"`
	Error         error         `json:"error,omitempty"`
}

// ReplayStats 回放统计
type ReplayStats struct {
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Duration       time.Duration `json:"duration"`
	TotalEvents    int           `json:"total_events"`
	ReplayedEvents int           `json:"replayed_events"`
	SkippedEvents  int           `json:"skipped_events"`
	ErrorEvents    int           `json:"error_events"`
	AverageDelay   time.Duration `json:"average_delay"`
	MinDelay       time.Duration `json:"min_delay"`
	MaxDelay       time.Duration `json:"max_delay"`
	PauseCount     int           `json:"pause_count"`
	TotalPauseTime time.Duration `json:"total_pause_time"`
}

// ReplayCallback 回放回调函数
type ReplayCallback func(event *ReplayEvent) error

// SessionReplayer 会话回放器
type SessionReplayer struct {
	session   *Session
	config    *ReplayConfig
	callbacks []ReplayCallback
	stats     *ReplayStats

	// 控制状态
	isPlaying   bool
	isPaused    bool
	currentTime time.Time

	// 同步控制
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	pauseCh  chan struct{}
	resumeCh chan struct{}
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// NewSessionReplayer 创建新的会话回放器
func NewSessionReplayer(session *Session, config *ReplayConfig) *SessionReplayer {
	if config == nil {
		config = &ReplayConfig{
			Speed:         SpeedNormal,
			EnablePause:   false,
			PauseOnError:  false,
			MaxReplayTime: 0,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	replayer := &SessionReplayer{
		session:   session,
		config:    config,
		callbacks: make([]ReplayCallback, 0),
		stats: &ReplayStats{
			StartTime: time.Now(),
		},
		ctx:      ctx,
		cancel:   cancel,
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
		stopCh:   make(chan struct{}),
	}

	return replayer
}

// AddCallback 添加回放回调
func (r *SessionReplayer) AddCallback(callback ReplayCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callbacks = append(r.callbacks, callback)
}

// Play 开始回放
func (r *SessionReplayer) Play() error {
	if r.isPlaying {
		return fmt.Errorf("replay is already playing")
	}

	r.mu.Lock()
	r.isPlaying = true
	r.isPaused = false
	r.currentTime = r.session.StartTime
	r.stats.StartTime = time.Now()
	r.mu.Unlock()

	// 启动回放协程
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer r.stopOnce.Do(func() { close(r.stopCh) })
		r.replayLoop()
	}()

	return nil
}

// Pause 暂停回放
func (r *SessionReplayer) Pause() error {
	if !r.isPlaying {
		return fmt.Errorf("replay is not playing")
	}

	if r.isPaused {
		return fmt.Errorf("replay is already paused")
	}

	r.mu.Lock()
	r.isPaused = true
	r.stats.PauseCount++
	r.mu.Unlock()

	// 发送暂停信号
	select {
	case r.pauseCh <- struct{}{}:
	default:
	}

	return nil
}

// Resume 恢复回放
func (r *SessionReplayer) Resume() error {
	if !r.isPlaying {
		return fmt.Errorf("replay is not playing")
	}

	if !r.isPaused {
		return fmt.Errorf("replay is not paused")
	}

	r.mu.Lock()
	r.isPaused = false
	r.mu.Unlock()

	// 发送恢复信号
	select {
	case r.resumeCh <- struct{}{}:
	default:
	}

	return nil
}

// Stop 停止回放
func (r *SessionReplayer) Stop() error {
	r.mu.Lock()
	if !r.isPlaying {
		r.mu.Unlock()
		return nil
	}
	
	r.isPlaying = false
	r.isPaused = false
	r.stats.EndTime = time.Now()
	r.stats.Duration = r.stats.EndTime.Sub(r.stats.StartTime)
	r.mu.Unlock()

	r.stopOnce.Do(func() { close(r.stopCh) })
	r.cancel()
	return nil
}

// Wait 等待回放完成
func (r *SessionReplayer) Wait() {
	r.wg.Wait()
}

// ReplayedEvents 获取已回放事件数量
func (r *SessionReplayer) ReplayedEvents() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stats.ReplayedEvents
}

// GetStats 获取回放统计
func (r *SessionReplayer) GetStats() *ReplayStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := *r.stats
	if r.isPlaying && !r.stats.EndTime.IsZero() {
		stats.Duration = time.Since(r.stats.StartTime)
	}

	return &stats
}

// replayLoop 回放主循环
func (r *SessionReplayer) replayLoop() {
	defer func() {
		r.mu.Lock()
		r.isPlaying = false
		r.stats.EndTime = time.Now()
		r.stats.Duration = r.stats.EndTime.Sub(r.stats.StartTime)
		r.mu.Unlock()
	}()

	// 按时间顺序排序事件
	events := r.sortEventsByTime()

	for _, event := range events {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		// 检查是否需要暂停
		r.mu.RLock()
		paused := r.isPaused
		r.mu.RUnlock()
		
		if paused {
			select {
			case <-r.resumeCh:
				r.mu.Lock()
				r.isPaused = false
				r.mu.Unlock()
			case <-r.ctx.Done():
				return
			case <-r.stopCh:
				return
			}
		}

		// 应用事件过滤器
		if !r.shouldReplayEvent(event) {
			r.mu.Lock()
			r.stats.SkippedEvents++
			r.mu.Unlock()
			continue
		}

		// 计算回放延迟
		delay := r.calculateReplayDelay(event)

		// 创建回放事件
		replayEvent := &ReplayEvent{
			OriginalEvent: event,
			ReplayTime:    time.Now(),
			Delay:         delay,
		}

		// 执行回调
		if err := r.executeCallbacks(replayEvent); err != nil {
			replayEvent.Error = err
			r.stats.ErrorEvents++

			if r.config.PauseOnError {
				r.Pause()
			}
		} else {
			r.mu.Lock()
			r.stats.ReplayedEvents++
			r.mu.Unlock()
		}

		// 更新统计
		r.updateReplayStats(replayEvent)

		// 等待延迟时间
		if r.config.Speed > 0 {
			waitTime := time.Duration(float64(delay) / float64(r.config.Speed))
			if waitTime > 0 {
				select {
				case <-time.After(waitTime):
				case <-r.ctx.Done():
					return
				}
			}
		}
	}
}

// sortEventsByTime 按时间排序事件
func (r *SessionReplayer) sortEventsByTime() []*SessionEvent {
	events := make([]*SessionEvent, len(r.session.Events))
	copy(events, r.session.Events)

	// 按时间戳排序
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.After(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return events
}

// shouldReplayEvent 检查是否应该回放事件
func (r *SessionReplayer) shouldReplayEvent(event *SessionEvent) bool {
	if len(r.config.EventFilter.EventTypes) > 0 {
		found := false
		for _, eventType := range r.config.EventFilter.EventTypes {
			if event.Type == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if r.config.EventFilter.Opcode != nil && event.Opcode != *r.config.EventFilter.Opcode {
		return false
	}

	return true
}

// calculateReplayDelay 计算回放延迟
func (r *SessionReplayer) calculateReplayDelay(event *SessionEvent) time.Duration {
	if r.currentTime.IsZero() {
		r.currentTime = event.Timestamp
		return 0
	}

	delay := event.Timestamp.Sub(r.currentTime)
	r.currentTime = event.Timestamp
	return delay
}

// executeCallbacks 执行回调函数
func (r *SessionReplayer) executeCallbacks(event *ReplayEvent) error {
	r.mu.RLock()
	callbacks := make([]ReplayCallback, len(r.callbacks))
	copy(callbacks, r.callbacks)
	r.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(event); err != nil {
			return err
		}
	}

	return nil
}

// updateReplayStats 更新回放统计
func (r *SessionReplayer) updateReplayStats(event *ReplayEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新延迟统计
	if event.Delay > 0 {
		if r.stats.MinDelay == 0 || event.Delay < r.stats.MinDelay {
			r.stats.MinDelay = event.Delay
		}
		if event.Delay > r.stats.MaxDelay {
			r.stats.MaxDelay = event.Delay
		}

		// 计算平均延迟
		totalDelay := r.stats.AverageDelay*time.Duration(r.stats.ReplayedEvents) + event.Delay
		r.stats.AverageDelay = totalDelay / time.Duration(r.stats.ReplayedEvents+1)
	}
}

// IsPlaying 检查是否正在回放
func (r *SessionReplayer) IsPlaying() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isPlaying
}

// IsPaused 检查是否已暂停
func (r *SessionReplayer) IsPaused() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isPaused
}

// GetCurrentTime 获取当前回放时间
func (r *SessionReplayer) GetCurrentTime() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentTime
}

// SeekTo 跳转到指定时间
func (r *SessionReplayer) SeekTo(targetTime time.Time) error {
	if r.isPlaying {
		return fmt.Errorf("cannot seek while playing")
	}

	r.mu.Lock()
	r.currentTime = targetTime
	r.mu.Unlock()

	return nil
}
