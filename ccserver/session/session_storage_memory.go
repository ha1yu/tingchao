package session

import (
	"errors"
	"log"
	"sync"
	"time"
)

// IStorage --------------session_from---------------------------
// session存储方式接口，可以存储在内存，数据库或者文件
// 分别实现该接口即可
// 如存入数据库的CRUD操作
type IStorage interface {
	InitSession(sessionID string, maxAge int64) (ISession, error) // 初始化一个session，id根据需要生成后传入
	SetSession(session ISession)                                  // 根据sid，获得当前session
	GetSession(sessionID string) ISession                         // 根据sessionID获取session,不存在返回nil
	RemoveSession(sessionID string) error                         // 销毁session
	GCSession()                                                   // session GC
}

// StorageFromMemory session存储在内存中
type StorageFromMemory struct {
	Lock       sync.RWMutex
	SessionMap map[string]ISession
}

func newStorageFromMemory() *StorageFromMemory {
	return &StorageFromMemory{
		SessionMap: make(map[string]ISession),
	}
}

// InitSession 初始化会话session，这个结构体操作实现Session接口
func (s *StorageFromMemory) InitSession(sessionID string, maxAge int64) (ISession, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	session := newSessionFromMemory()

	session.SetId(sessionID)                // 设置sessionID
	session.SetMaxAge(maxAge)               // 设置最大有效时间
	session.SetLastAccessedTime(time.Now()) // 设置上次访问时间

	s.SessionMap[session.GetId()] = session
	return session, nil
}

func (s *StorageFromMemory) SetSession(session ISession) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.SessionMap[session.GetId()] = session
}

func (s *StorageFromMemory) GetSession(sessionID string) ISession {
	s.Lock.RLock()
	defer s.Lock.RUnlock()

	if session, ok := s.SessionMap[sessionID]; ok { // 如果有session返回,否则返回nil
		return session
	}
	return nil
}

// RemoveSession 销毁session
func (s *StorageFromMemory) RemoveSession(sessionID string) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if _, ok := s.SessionMap[sessionID]; ok {
		delete(s.SessionMap, sessionID)
		return nil
	}
	return errors.New("RemoveSession error")
}

// GCSession 监判超时
func (s *StorageFromMemory) GCSession() {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	sessionMap := s.SessionMap
	if len(sessionMap) < 1 { // sessionMap中没有session直接返回,无需GC
		return
	}

	time1 := time.Now().UnixMilli() // 老版本GO使用此cpi会报错,老版本GO time类不存在获取毫秒的方法(UnixMilli)
	num := 0
	for name, sess := range sessionMap { // 遍历所有session,删除超时的session对象,清理内存占用
		// times := (sess.(*Session).lastAccessedTime.Unix()) + (sess.(*Session).maxAge)
		times := (sess.(*Session).GetLastAccessedTime().Unix()) + (sess.(*Session).GetMaxAge())
		if times < time.Now().Unix() { // session超时之后则删除此session
			num = num + 1
			delete(s.SessionMap, name)
		}
	}
	time2 := time.Now().UnixMilli()
	time3 := time2 - time1
	log.Println("[session.GC]本次GC耗时", time3, "毫秒;", "删除失效session", num, "个;",
		"系统session剩余总数为", len(s.SessionMap))
}
