package session

import (
	"time"
)

// Session 完整的会话记录
type Session struct {
	ID        string           `json:"id"`
	StartTime time.Time        `json:"start_time"`
	EndTime   time.Time        `json:"end_time"`
	Events    []*SessionEvent  `json:"events"`
	Frames    []*MessageFrame  `json:"frames"`
	Stats     *SessionStats    `json:"stats"`
} 