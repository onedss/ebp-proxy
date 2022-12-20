package proxy

import (
	"fmt"
	"github.com/onedss/ebp-proxy/core"
	"github.com/onedss/ebp-proxy/mytool"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Pusher struct {
	core.SessionLogger
	*Session

	Path    string
	Stopped bool

	cond  *sync.Cond
	queue []*DataPack
}

func NewPusher(path string, session *Session) (pusher *Pusher) {
	pusher = &Pusher{
		Session: session,
		Path:    path,
		cond:    sync.NewCond(&sync.Mutex{}),
		queue:   make([]*DataPack, 0),
	}
	pusher.SetLogger(log.New(os.Stdout, fmt.Sprintf("[%s] ", session.ID), log.LstdFlags|log.Lshortfile|log.Lmicroseconds))
	if !mytool.Debug {
		pusher.GetLogger().SetOutput(mytool.GetLogWriter())
	}
	session.AddRTPHandles(func(pack *DataPack) {
		pusher.SendPack(pack)
	})
	session.AddStopHandles(func() {
		pusher.Server.RemovePusher(pusher)
		pusher.cond.Broadcast()
	})
	return
}

func (pusher *Pusher) GetPath() string {
	return pusher.Path
}

func (pusher *Pusher) GetID() string {
	return pusher.ID
}

func (pusher *Pusher) SendPack(pack *DataPack) *Pusher {
	var err error
	logger := pusher.GetLogger()
	_, err = net.ResolveTCPAddr("tcp", pusher.GetPath())
	if err != nil {
		logger.Printf("Pusher[%s] 网络地址错误. [%s]", pusher.Path, pusher.ID)
		return pusher
	}
	var conn net.Conn
	hostPort := pusher.Path
	if conn, err = net.DialTimeout("tcp4", hostPort, 500*time.Millisecond); err != nil {
		logger.Printf("Pusher[%s] 发送数据超时. [%s], 错误: %v", hostPort, pusher.ID, err)
		conn.Close()
		return pusher
	}
	data := pack.Buffer.Bytes()
	if conn != nil {
		conn.Write(data)
		conn.Close()
	}
	logger.Printf("%v", data)
	return pusher
}

func (pusher *Pusher) QueuePack(pack *DataPack) *Pusher {
	pusher.cond.L.Lock()
	pusher.queue = append(pusher.queue, pack)
	pusher.cond.Signal()
	pusher.cond.L.Unlock()
	return pusher
}

func (pusher *Pusher) StartPush() {
	logger := pusher.GetLogger()
	logger.Printf("Pusher[%s] StartPush() Begin. [%s]", pusher.Path, pusher.ID)
	for !pusher.Stopped {
		var pack *DataPack
		pusher.cond.L.Lock()
		if len(pusher.queue) == 0 {
			pusher.cond.Wait()
		}
		if len(pusher.queue) > 0 {
			pack = pusher.queue[0]
			pusher.queue = pusher.queue[1:]
		}
		pusher.cond.L.Unlock()
		if pack == nil {
			if !pusher.Stopped {
				logger.Printf("Pusher[%s] not stopped, but queue take out nil pack", pusher.Path)
			}
			continue
		}
		pusher.BroadcastRTP(pack)
	}
	logger.Printf("Pusher[%s] StartPush() End. [%s]", pusher.Path, pusher.ID)
}

func (pusher *Pusher) StopPush() {
	pusher.Stopped = true
}

func (pusher *Pusher) BroadcastRTP(pack *DataPack) *Pusher {
	logger := pusher.GetLogger()
	logger.Printf("Pusher[%s] BroadcastRTP(). [%s]", pusher.Path, pusher.ID)
	return pusher
}
