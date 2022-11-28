package app

import "net/http"

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
	return nil
}

func (p *httpServer) Stop() (err error) {
	p.httpStop()
	return nil
}

func (p *httpServer) GetPort() int {
	return p.httpPort
}

func (p *httpServer) httpStop() (err error) {
	return nil
}
