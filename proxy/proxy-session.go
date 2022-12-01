package proxy

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/onedss/ebp-proxy/core"
	"github.com/onedss/ebp-proxy/mylog"
	"github.com/teris-io/shortid"
	"io"
	"net"
	"sync"
	"time"
)

type Session struct {
	core.SessionLogger
	ID          string
	Server      *Server
	privateConn *RichConn
	connRW      *bufio.ReadWriter
	connWLock   sync.RWMutex

	// stats info
	InBytes  int
	OutBytes int
	StartAt  time.Time
	Timeout  int

	Stopped bool

	RTPHandles  []func(buffer *bytes.Buffer)
	StopHandles []func()
}

func NewSession(server *Server, conn *net.TCPConn) *Session {
	networkBuffer := mylog.Conf().Section("proxy").Key("network_buffer").MustInt(204800)
	timeoutMillis := mylog.Conf().Section("proxy").Key("timeout").MustInt(0)
	timeoutTCPConn := &RichConn{conn, time.Duration(timeoutMillis) * time.Millisecond}
	session := &Session{
		ID:          shortid.MustGenerate(),
		Server:      server,
		privateConn: timeoutTCPConn,
		connRW:      bufio.NewReadWriter(bufio.NewReaderSize(timeoutTCPConn, networkBuffer), bufio.NewWriterSize(timeoutTCPConn, networkBuffer)),
		StartAt:     time.Now(),
		Timeout:     timeoutMillis,

		RTPHandles:  make([]func(*bytes.Buffer), 0),
		StopHandles: make([]func(), 0),
	}
	if !mylog.Debug {
		session.GetLogger().SetOutput(mylog.GetLogWriter())
	}
	return session
}

func (session *Session) AddRTPHandles(f func(*bytes.Buffer)) {
	session.RTPHandles = append(session.RTPHandles, f)
}

func (session *Session) AddStopHandles(f func()) {
	session.StopHandles = append(session.StopHandles, f)
}

func (session *Session) Stop() {
	logger := session.GetLogger()
	logger.Printf("Session Stop. [%s] First Close: %v", session.ID, session.Stopped == false)
	if session.Stopped {
		return
	}
	session.Stopped = true
	for _, h := range session.StopHandles {
		h()
	}
	if session.privateConn != nil {
		session.connRW.Flush()
		session.privateConn.Close()
		session.privateConn = nil
	}
}

func (session *Session) Start() {
	defer session.Stop()
	bufHead := make([]byte, 2)
	bufVer := make([]byte, 2)
	bufSessionId := make([]byte, 4)
	bufType := make([]byte, 1)
	bufLen := make([]byte, 2)
	logger := session.GetLogger()
	logger.Printf("Session Start. [%s]", session.ID)
	for !session.Stopped {
		if _, err := io.ReadFull(session.connRW, bufHead); err != nil {
			logger.Println(session, err)
			return
		}
		if bufHead[0] != 0xFE || bufHead[1] != 0xFD {
			logger.Println("数据包头标志错误：", session.privateConn.RemoteAddr())
			break
		}
		if _, err := io.ReadFull(session.connRW, bufVer); err != nil {
			logger.Println(err)
			return
		}
		if bufVer[0] != 0x01 || bufVer[1] != 0x00 {
			logger.Println("数据协议版本号错误：", session.privateConn.RemoteAddr())
			break
		}
		if _, err := io.ReadFull(session.connRW, bufSessionId); err != nil {
			logger.Println(err)
			return
		}
		if _, err := io.ReadFull(session.connRW, bufType); err != nil {
			logger.Println(err)
			return
		}
		if _, err := io.ReadFull(session.connRW, bufLen); err != nil {
			logger.Println(err)
			return
		}
		pkgLen := int(binary.BigEndian.Uint16(bufLen))
		if pkgLen < 11 || pkgLen > 1024 {
			logger.Println("数据包长度错误：", pkgLen, session.privateConn.RemoteAddr())
			break
		}
		bufBody := make([]byte, pkgLen-11)
		if _, err := io.ReadFull(session.connRW, bufBody); err != nil {
			logger.Println(err)
			return
		}
		reqBuf := bytes.NewBuffer(nil)
		reqBuf.Write(bufHead)
		reqBuf.Write(bufVer)
		reqBuf.Write(bufSessionId)
		reqBuf.Write(bufType)
		reqBuf.Write(bufLen)
		reqBuf.Write(bufBody)
		for _, h := range session.RTPHandles {
			h(reqBuf)
		}
		//data := reqBuf.Bytes()
		//os.Stdout.Write(data)
		logger.Println("正常读完数据后关闭连接。", session.privateConn.RemoteAddr(), "数据包长度:", pkgLen)
		break
	}
}
