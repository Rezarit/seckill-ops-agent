// Package mem 提供简单的内存存储功能
// 用于管理对话历史记录，支持多会话隔离
package mem

import (
	"sync"

	"github.com/cloudwego/eino/schema"
)

// SimpleMemoryMap 存储所有会话的内存实例
// key: 会话ID, value: 会话内存实例
var SimpleMemoryMap = make(map[string]*SimpleMemory)

// mu 用于保护 SimpleMemoryMap 的并发访问
var mu sync.Mutex

// GetSimpleMemory 获取指定会话的内存实例（线程安全）
// 如果会话不存在，则创建新的内存实例
// 参数:
//   - id: 会话ID
// 返回: 会话内存实例
func GetSimpleMemory(id string) *SimpleMemory {
	mu.Lock()
	defer mu.Unlock()

	// 如果存在就返回，不存在就创建
	if mem, ok := SimpleMemoryMap[id]; ok {
		return mem
	}

	// 创建新的内存实例，默认最大窗口大小为6
	newMem := &SimpleMemory{
		ID:            id,
		Messages:      []*schema.Message{},
		MaxWindowSize: 6,
	}
	SimpleMemoryMap[id] = newMem
	return newMem
}

// SimpleMemory 简单的对话内存实现
// 维护对话历史记录，支持自动窗口大小限制
type SimpleMemory struct {
	ID            string            `json:"id"`             // 会话ID
	Messages      []*schema.Message `json:"messages"`       // 消息列表（按时间顺序）
	MaxWindowSize int               // 最大窗口大小（消息数量上限）
	mu            sync.Mutex        // 保护消息列表的并发访问
}

// SetMessages 添加消息到内存
// 自动维护窗口大小，超出时丢弃最早的消息对（保持对话配对关系）
// 参数:
//   - msg: 要添加的消息
func (c *SimpleMemory) SetMessages(msg *schema.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 添加新消息
	c.Messages = append(c.Messages, msg)

	// 如果超过最大窗口大小，需要丢弃旧消息
	if len(c.Messages) > c.MaxWindowSize {
		// 确保成对丢弃消息，保持用户-助手对话配对关系
		excess := len(c.Messages) - c.MaxWindowSize
		if excess%2 != 0 {
			excess++ // 确保丢弃偶数条消息
		}
		// 丢弃前面的消息，保持最新的对话
		c.Messages = c.Messages[excess:]
	}
}

// GetMessages 获取当前会话的所有消息
// 返回: 消息列表的副本（线程安全）
func (c *SimpleMemory) GetMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Messages
}
