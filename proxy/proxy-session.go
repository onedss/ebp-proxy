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

	connReader *bufio.Reader
	connWriter *bufio.Writer

	// stats info
	InBytes  int
	OutBytes int
	StartAt  time.Time
	Timeout  int

	Stopped bool

	Pusher *Pusher

	DataHandles []func(*DataPack)
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
		//connRW:      bufio.NewReadWriter(bufio.NewReaderSize(timeoutTCPConn, networkBuffer), bufio.NewWriterSize(timeoutTCPConn, networkBuffer)),
		StartAt: time.Now(),
		Timeout: timeoutMillis,

		DataHandles: make([]func(*DataPack), 0),
		StopHandles: make([]func(), 0),
	}
	session.connReader = bufio.NewReaderSize(timeoutTCPConn, networkBuffer)
	session.connWriter = bufio.NewWriterSize(timeoutTCPConn, networkBuffer)
	session.connRW = bufio.NewReadWriter(session.connReader, session.connWriter)
	if !mylog.Debug {
		session.GetLogger().SetOutput(mylog.GetLogWriter())
	}
	target1 := mylog.Conf().Section("proxy").Key("target1").MustString("")
	session.AddPusher(target1)
	target2 := mylog.Conf().Section("proxy").Key("target2").MustString("")
	session.AddPusher(target2)
	target3 := mylog.Conf().Section("proxy").Key("target3").MustString("")
	session.AddPusher(target3)
	target4 := mylog.Conf().Section("proxy").Key("target4").MustString("")
	session.AddPusher(target4)
	target5 := mylog.Conf().Section("proxy").Key("target5").MustString("")
	session.AddPusher(target5)
	return session
}

func (session *Session) AddPusher(path string) {
	if path == "" {
		return
	}
	session.Pusher = NewPusher(path, session)
	if session.Server.GetPusher(path) == nil {
		session.Server.AddPusher(session.Pusher)
	}
}

func (session *Session) AddRTPHandles(f func(*DataPack)) {
	session.DataHandles = append(session.DataHandles, f)
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
		session.connRW = nil
		session.connWriter = nil
		session.connReader = nil
		logger.Printf("正常读完数据后释放连接。[%s]", session.ID)
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
		var pack *DataPack
		if _, err := io.ReadFull(session.connRW, bufHead); err != nil {
			//logger.Println(session, err)
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
		pack = &DataPack{
			Type:   GDJ0892018E,
			Buffer: reqBuf,
		}
		for _, h := range session.DataHandles {
			h(pack)
		}
		logger.Println("正常读完数据后关闭连接。", session.privateConn.RemoteAddr(), "数据包长度:", pkgLen)
		break
	}
}
