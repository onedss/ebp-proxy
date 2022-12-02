package proxy

import (
	"fmt"
	"github.com/onedss/ebp-proxy/core"
	"github.com/onedss/ebp-proxy/mylog"
	"net"
	"sync"
)

type Server struct {
	core.SessionLogger
	TCPListener *net.TCPListener
	TCPPort     int
	Stopped     bool

	pushers     map[string]*Pusher // Path <-> Pusher
	pushersLock sync.RWMutex
	//addPusherCh    chan *Pusher
	//removePusherCh chan *Pusher

	networkBuffer int
}

func NewOneProxyServer(proxyPort int) (server *Server) {
	networkBuffer := mylog.Conf().Section("proxy").Key("network_buffer").MustInt(1048576)
	return &Server{
		SessionLogger: core.NewSessionLogger("[ProxyServer] "),
		TCPPort:       proxyPort,
		pushers:       make(map[string]*Pusher),
		//addPusherCh:    make(chan *Pusher),
		//removePusherCh: make(chan *Pusher),
		networkBuffer: networkBuffer,
	}
}

func (server *Server) Start() (err error) {
	var (
		addr     *net.TCPAddr
		listener *net.TCPListener
	)
	logger := server.GetLogger()
	addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", server.TCPPort))
	if err != nil {
		return
	}
	listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}
	server.Stopped = false
	server.TCPListener = listener
	logger.Println("Proxy server start on", server.TCPPort)
	for !server.Stopped {
		conn, err := server.TCPListener.AcceptTCP()
		if err != nil {
			logger.Println(err)
			continue
		}
		if err := conn.SetReadBuffer(server.networkBuffer); err != nil {
			logger.Printf("Proxy server conn set read buffer error, %v", err)
		}
		if err := conn.SetWriteBuffer(server.networkBuffer); err != nil {
			logger.Printf("Proxy server conn set write buffer error, %v", err)
		}
		session := NewSession(server, conn)
		go session.Start()
	}
	return nil
}

func (server *Server) Stop() (err error) {
	logger := server.GetLogger()
	logger.Println("Proxy server stop on", server.TCPPort)
	server.Stopped = true
	if server.TCPListener != nil {
		server.TCPListener.Close()
		server.TCPListener = nil
	}
	return nil
}

func (server *Server) GetPort() int {
	return server.TCPPort
}

func (server *Server) AddPusher(pusher *Pusher) {
	logger := server.GetLogger()
	server.pushersLock.Lock()
	if _, ok := server.pushers[pusher.GetPath()]; !ok {
		server.pushers[pusher.GetPath()] = pusher
		//go pusher.StartPush()
		logger.Printf("Pusher[%s] start, now pusher size[%d]", pusher.GetPath(), len(server.pushers))
	}
	server.pushersLock.Unlock()
}

func (server *Server) RemovePusher(pusher *Pusher) {
	logger := server.GetLogger()
	server.pushersLock.Lock()
	if _pusher, ok := server.pushers[pusher.GetPath()]; ok && pusher.GetID() == _pusher.GetID() {
		delete(server.pushers, pusher.GetPath())
		go pusher.StopPush()
		logger.Printf("Pusher[%s] end, now pusher size[%d]\n", pusher.GetPath(), len(server.pushers))
	}
	server.pushersLock.Unlock()
}

func (server *Server) GetPusher(path string) (pusher *Pusher) {
	server.pushersLock.RLock()
	pusher = server.pushers[path]
	server.pushersLock.RUnlock()
	return
}
