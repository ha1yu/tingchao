package session

import (
	"sync"
	"time"
)

// ISession -------------session_implements-----------------
// Session操作接口，不同存储方式的Session操作不同，实现也不同
type ISession interface {
	Set(sessionID, value interface{})      // 添加数据
	Get(sessionID interface{}) interface{} // 获取数据
	Remove(sessionID interface{}) bool     // 删除数据
	GetId() string                         // 获取sessionID
	SetId(sessionID string)                // 设置sessionID
	GetLastAccessedTime() time.Time        // 获取session最后访问时间
	SetLastAccessedTime(t time.Time)       // 设置session最后访问时间
	GetMaxAge() int64                      // 获取session超时时间
	SetMaxAge(maxAge int64)                // 设置session超时时间
}

// Session session实现
type Session struct {
	sessionID        string                      // ID不能相同
	lastAccessedTime time.Time                   // 最后访问时间
	maxAge           int64                       // 超时时间
	data             map[interface{}]interface{} // 主数据
	lock             sync.RWMutex                // 锁
}

//实例化
func newSessionFromMemory() *Session {
	return &Session{
		data:   make(map[interface{}]interface{}),
		maxAge: 86400, // 默认1天
	}
}

// Set 同一个会话均可调用，进行设置，改操作必须拥有排斥锁
func (s *Session) Set(sessionID, value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[sessionID] = value
}

func (s *Session) Get(sessionID interface{}) interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if value := s.data[sessionID]; value != nil {
		return value
	} else {
		return nil
	}
}

func (s *Session) Remove(sessionID interface{}) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if value := s.data[sessionID]; value != nil {
		delete(s.data, sessionID)
		return true
	} else {
		return false
	}
}

func (s *Session) GetId() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.sessionID
}

func (s *Session) SetId(sessionID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.sessionID = sessionID
}

func (s *Session) GetLastAccessedTime() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.lastAccessedTime
}

func (s *Session) SetLastAccessedTime(t time.Time) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.lastAccessedTime = t
}

func (s *Session) GetMaxAge() int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.maxAge
}

func (s *Session) SetMaxAge(maxAge int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.maxAge = maxAge
}
