package session

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/titan/tingchao/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type ISessionManager interface {
	GetCookieName() string                                      // 获取cookie名
	SetCookieName(cookieName string)                            // 设置cookie名
	GetSession(r *http.Request, w http.ResponseWriter) ISession // 无条件获取一个session,没有会生成一个
	UpdateSessionLastAccessedTime(r *http.Request) bool         // 更新session最后访问时间
	UpdateSessionLastAccessedTime1(session ISession)            // 更新session最后访问时间
	GetSessionById(sessionID string) ISession                   // 通过sessionID获取session,找不到返回nil值
	RemoveSession(w http.ResponseWriter, r *http.Request)       // 手动销毁session，同时删除cookie
	GC()                                                        // 根据按照固定的时间执行GC
	SetMaxAge(t int64)                                          // 设置session和cookie最大有效时间
	randomId() string                                           // 随机生成一个sessionID
}

// SessionManager --------------session_manager----------------------
// 管理Session,实际操作cookie，Storage
// 由于该结构体是整个应用级别的，写、修改都需要枷锁
type SessionManager struct {
	lock       sync.RWMutex // 锁
	cookieName string       // session key值,存放在客户端
	Storage    IStorage     // 存放方式，如内存，数据库，文件等
	maxAge     int64        // 超时时间,GC中使用
}

// NewSessionManager 实例化一个session管理器
func NewSessionManager() *SessionManager {
	sessionManager := &SessionManager{
		cookieName: utils.GlobalConfig.SessionCookieName,
		Storage:    newStorageFromMemory(),         // 默认使用内存版本的
		maxAge:     utils.GlobalConfig.SessionTime, // 多久执行一次session GC,单位:秒
	}
	go sessionManager.GC() // 开一个go程后台执行GC,经过测试20万Map数据GC需要90毫秒左右

	return sessionManager
}

func (m *SessionManager) GetCookieName() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.cookieName
}

func (m *SessionManager) SetCookieName(cookieName string) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	m.cookieName = cookieName
}

// GetSession 无条件获取一个session,没有会生成一个
// 先判断当前请求的cookie中是否存在有效的session,存在返回，不存在创建
// 客户端携带cookie,先检测cookie对应的session是否存在,存在返回session,
// 不存在则新创建一个session,并返回一个新cookie
func (m *SessionManager) GetSession(r *http.Request, w http.ResponseWriter) ISession {
	m.lock.Lock() // 锁住 cookieName,maxAge 等属性
	defer m.lock.Unlock()

	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" { // 客户端未携带cooke,则创建一个cookie给客户端

		sessionID := m.randomId()                                // 随机创建一个不同的sessionID
		session, _ := m.Storage.InitSession(sessionID, m.maxAge) //根据保存session方式,如内存,数据库中创建
		maxAge := m.maxAge                                       // 设置session最大有效时间

		if maxAge == 0 {
			maxAge = 86400 // session默认有效时间
		}

		// 用session的ID于cookie关联 cookie名字和失效时间由session管理器维护
		cookie := http.Cookie{
			Name:     m.cookieName,
			Value:    url.QueryEscape(sessionID), // 转义特殊符号@#￥%+*-等
			Path:     "/",
			HttpOnly: true,
			MaxAge:   int(maxAge),
			Expires:  time.Now().Add(time.Duration(maxAge)),
		}
		http.SetCookie(w, &cookie)
		return session

	} else { // 客户端携带cookie,先检测cookie对应的session是否存在,存在返回session,不存在则新创建一个session,并返回一个新cookie

		sessionID, _ := url.QueryUnescape(cookie.Value) // 反转义特殊符号
		session := m.Storage.GetSession(sessionID)      // 从保存session介质中获取session
		if session == nil {                             // session不存在则新创建一个session,并返回一个新cookie
			//log.Println("cookie存在，但是session已经被清理")

			sessionID := m.randomId()                                   // 没有session则代表被GC清理了,在创建一个
			newSession, _ := m.Storage.InitSession(sessionID, m.maxAge) // 添加一个session
			maxAge := m.maxAge                                          // 设置session最大有效时间

			if maxAge == 0 {
				maxAge = 86400 // session默认有效时间
			}

			// 用session的ID于cookie关联 cookie名字和失效时间由session管理器维护
			newCookie := http.Cookie{
				Name:     m.cookieName,
				Value:    url.QueryEscape(sessionID), //转义特殊符号@#￥%+*-等
				Path:     "/",
				HttpOnly: true,
				MaxAge:   int(maxAge),
				Expires:  time.Now().Add(time.Duration(maxAge)),
			}
			http.SetCookie(w, &newCookie)
			return newSession
		}
		return session // 找到session,则直接返回session
	}

}

// UpdateSessionLastAccessedTime 更新最后访问时间,返回是否更新成功
func (m *SessionManager) UpdateSessionLastAccessedTime(req *http.Request) bool {

	cookie, err := req.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" { // 客户端未携带cooke,无法完成更新动作,直接返回
		return false
	}

	sessionID, _ := url.QueryUnescape(cookie.Value) // 反转义特殊符号
	session := m.Storage.GetSession(sessionID)      // 从保存session介质中获取session
	if session == nil {                             // session不存在,可能是已经被GC了,无法完成更新动作,直接返回
		return false
	}

	t := time.Now()
	session.SetLastAccessedTime(t) // 更新session 最后访问时间
	m.Storage.SetSession(session)  // 更新session

	// 暂时不需要更新客户端cookie,以下代码可以用在web浏览器中,不适合客户端
	/*
		if m.maxAge != 0 {
			cookie.MaxAge = int(m.maxAge)
		} else {
			cookie.MaxAge = int(session.(*Session).maxAge)
		}
		http.SetCookie(w, cookie) // 更新客户端cookie有效期
	*/

	return true
}

// UpdateSessionLastAccessedTime1 更新最后访问时间
func (m *SessionManager) UpdateSessionLastAccessedTime1(session ISession) {
	t := time.Now()
	session.SetLastAccessedTime(t) // 更新session 最后访问时间
	m.Storage.SetSession(session)  // 更新session
}

// GetSessionById 通过sessionID获取session,找不到返回nil值
func (m *SessionManager) GetSessionById(sessionID string) ISession {
	sessionID, err := url.QueryUnescape(sessionID) // 将url转码过的字符串再转回来
	if err != nil {
		log.Println("[session.SessionManager.GetSessionById]url解码出现错误", err)
		return nil
	}
	session := m.Storage.GetSession(sessionID)
	return session
}

// RemoveSession 手动销毁session，同时删除cookie
func (m *SessionManager) RemoveSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		sessionID, _ := url.QueryUnescape(cookie.Value)
		m.Storage.RemoveSession(sessionID)

		cookie2 := http.Cookie{
			MaxAge:  0,
			Name:    m.cookieName,
			Value:   "",
			Path:    "/",
			Expires: time.Now().Add(time.Duration(0)),
		}

		http.SetCookie(w, &cookie2)
	}
}

// GC 开启每个会话，同时定时调用该方法
// 到达session最大生命时，且超时时。回收它
func (m *SessionManager) GC() {
	ticker := time.NewTicker(time.Duration(utils.GlobalConfig.SessionGcTime) * time.Second)
	go func() {
		for {
			<-ticker.C
			m.Storage.GCSession()
		}
	}()
	//在多长时间后执行匿名函数，这里指在某个时间后执行GC
	//time.AfterFunc(time.Duration(m.maxAge/2), func() {
	// time.AfterFunc(time.Duration(utils.GlobalConfig.SessionGcTime)*time.Second, func() {
	// 	m.GC()
	// })
}

func (m *SessionManager) SetMaxAge(t int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.maxAge = t
}

// 生成一定长度的随机数
func (m *SessionManager) randomId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil { // 生成一个32位二进制随机数
		return ""
	}
	return base64.URLEncoding.EncodeToString(b) // url编码
}
