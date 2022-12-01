package app

import (
	"fmt"
	"github.com/onedss/ebp-proxy/models"
	"github.com/onedss/ebp-proxy/mylog"
	"github.com/onedss/ebp-proxy/routers"
	"github.com/onedss/ebp-proxy/service"
	"log"
)

type application struct {
	servers []OneServer
}

func (p *application) Start(s service.Service) (err error) {
	log.Println("********** START **********")
	for _, server := range p.servers {
		port := server.GetPort()
		if mylog.IsPortInUse(port) {
			err = fmt.Errorf("TCP port[%d] In Use", port)
			return
		}
	}
	err = models.Init()
	if err != nil {
		return
	}
	err = routers.Init()
	if err != nil {
		return
	}
	for _, server := range p.servers {
		go func(s OneServer) {
			if err := s.Start(); err != nil {
				log.Println("The server error!", err)
			}
			log.Println("The server is end. port:", s.GetPort())
		}(server)
	}
	go func() {
		for range routers.API.RestartChan {
			log.Println("********** STOP **********")
			for _, server := range p.servers {
				server.Stop()
			}
			mylog.ReloadConf()
			log.Println("********** START **********")
			for _, server := range p.servers {
				err := server.Start()
				if err != nil {
					return
				}
			}
		}
	}()
	return nil
}

func (p *application) Stop(s service.Service) (err error) {
	defer log.Println("********** STOP **********")
	defer mylog.CloseLogWriter()
	for _, server := range p.servers {
		server.Stop()
	}
	models.Close()
	return
}

func (p *application) AddServer(server OneServer) {
	p.servers = append(p.servers, server)
}
