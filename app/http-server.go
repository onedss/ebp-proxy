package app

import (
	"context"
	"fmt"
	"github.com/onedss/ebp-proxy/mylog"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

type httpServer struct {
	httpPort   int
	httpServer *http.Server
}

func NewOneHttpServer(httpPort int) (server *httpServer) {
	return &httpServer{
		httpPort: httpPort,
	}
}

func (p *httpServer) Start() (err error) {
	p.httpServer = &http.Server{
		Addr: fmt.Sprintf(":%d", p.httpPort),
		//Handler:           routers.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/status", myHandler)
	link := fmt.Sprintf("http://%s:%d", mylog.LocalIP(), p.httpPort)
	log.Println("Start http server -->", link)
	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Start http server error", err)
		}
		log.Println("Start http server end")
	}()
	return
}

func (p *httpServer) Stop() (err error) {
	if p.httpServer == nil {
		err = fmt.Errorf("HTTP Server Not Found")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = p.httpServer.Shutdown(ctx); err != nil {
		return
	}
	return
}

func (p *httpServer) GetPort() int {
	return p.httpPort
}

func (p *httpServer) httpStop() (err error) {
	return nil
}

// handler函数
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr, "连接成功")
	// 请求方式：GET POST DELETE PUT UPDATE
	fmt.Println("method:", r.Method)
	// /go
	fmt.Println("url:", r.URL.Path)
	fmt.Println("header:", r.Header)
	fmt.Println("body:", r.Body)
	// 回复
	w.Write([]byte("Welcome"))
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr, "连接成功")
	// 请求方式：GET POST DELETE PUT UPDATE
	fmt.Println("method:", r.Method)
	// /go
	fmt.Println("url:", r.URL.Path)
	fmt.Println("header:", r.Header)
	fmt.Println("body:", r.Body)
	// 回复
	w.Write([]byte("OK"))
}
