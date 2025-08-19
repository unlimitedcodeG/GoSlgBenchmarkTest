package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// 帧头长度：操作码(2字节) + 数据长度(4字节)
	FrameHeaderSize = 6
	// 最大帧大小限制（防止内存攻击）
	MaxFrameSize = 1024 * 1024 // 1MB
	// 最小帧大小（只有头部）
	MinFrameSize = FrameHeaderSize
)

var (
	ErrFrameTooSmall = errors.New("frame too small")
	ErrFrameTooLarge = errors.New("frame too large")
	ErrInvalidFrame  = errors.New("invalid frame format")
)

// Frame 表示一个完整的协议帧
type Frame struct {
	Opcode uint16 // 操作码
	Body   []byte // 消息体（protobuf序列化后的数据）
}

// EncodeFrame 将操作码和消息体编码为二进制帧格式
// 帧格式: | opcode(2字节) | length(4字节) | body(变长) |
func EncodeFrame(opcode uint16, body []byte) []byte {
	if body == nil {
		body = []byte{}
	}

	frameSize := FrameHeaderSize + len(body)
	buf := make([]byte, frameSize)

	// 写入操作码（大端序）
	binary.BigEndian.PutUint16(buf[0:2], opcode)
	// 写入消息体长度（大端序）
	binary.BigEndian.PutUint32(buf[2:6], uint32(len(body)))
	// 写入消息体
	copy(buf[6:], body)

	return buf
}

// DecodeFrame 从二进制数据中解码出操作码和消息体
func DecodeFrame(raw []byte) (opcode uint16, body []byte, err error) {
	if len(raw) < MinFrameSize {
		return 0, nil, ErrFrameTooSmall
	}

	if len(raw) > MaxFrameSize {
		return 0, nil, ErrFrameTooLarge
	}

	// 读取操作码
	opcode = binary.BigEndian.Uint16(raw[0:2])
	// 读取消息体长度
	bodyLength := binary.BigEndian.Uint32(raw[2:6])

	// 验证帧完整性
	expectedFrameSize := FrameHeaderSize + int(bodyLength)
	if len(raw) != expectedFrameSize {
		return 0, nil, fmt.Errorf("%w: expected %d bytes, got %d",
			ErrInvalidFrame, expectedFrameSize, len(raw))
	}

	// 提取消息体
	if bodyLength > 0 {
		body = make([]byte, bodyLength)
		copy(body, raw[6:])
	}

	return opcode, body, nil
}

// DecodeFrameFromReader 从数据流中逐步解码帧（用于流式读取）
type FrameDecoder struct {
	buffer     []byte
	headerRead bool
	frameSize  int
}

// NewFrameDecoder 创建新的帧解码器
func NewFrameDecoder() *FrameDecoder {
	return &FrameDecoder{
		buffer: make([]byte, 0, 1024),
	}
}

// Feed 向解码器输入数据
func (fd *FrameDecoder) Feed(data []byte) {
	fd.buffer = append(fd.buffer, data...)
}

// Next 尝试解码下一个完整的帧
func (fd *FrameDecoder) Next() (frame *Frame, err error) {
	// 如果还没有读取完整的头部
	if !fd.headerRead {
		if len(fd.buffer) < FrameHeaderSize {
			return nil, nil // 需要更多数据
		}

		// 读取帧头信息
		_ = binary.BigEndian.Uint16(fd.buffer[0:2]) // opcode，这里不需要使用
		bodyLength := binary.BigEndian.Uint32(fd.buffer[2:6])

		fd.frameSize = FrameHeaderSize + int(bodyLength)
		if fd.frameSize > MaxFrameSize {
			return nil, ErrFrameTooLarge
		}

		fd.headerRead = true
	}

	// 检查是否有完整的帧
	if len(fd.buffer) < fd.frameSize {
		return nil, nil // 需要更多数据
	}

	// 解码完整的帧
	frameData := fd.buffer[:fd.frameSize]
	opcode, body, err := DecodeFrame(frameData)
	if err != nil {
		return nil, err
	}

	frame = &Frame{
		Opcode: opcode,
		Body:   body,
	}

	// 移除已处理的数据
	fd.buffer = fd.buffer[fd.frameSize:]
	fd.headerRead = false
	fd.frameSize = 0

	return frame, nil
}

// Reset 重置解码器状态
func (fd *FrameDecoder) Reset() {
	fd.buffer = fd.buffer[:0]
	fd.headerRead = false
	fd.frameSize = 0
}

// BufferSize 返回当前缓冲区大小
func (fd *FrameDecoder) BufferSize() int {
	return len(fd.buffer)
}
