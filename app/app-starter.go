package app

import (
	"github.com/common-nighthawk/go-figure"
	"github.com/onedss/ebp-proxy/mylog"
	"github.com/onedss/ebp-proxy/proxy"
	"github.com/onedss/ebp-proxy/service"
	"log"
	"os"
)

func StartApp() {
	log.Println("ConfigFile -->", mylog.ConfFile())
	sec := mylog.Conf().Section("service")
	svcConfig := &service.Config{
		Name:        sec.Key("name").MustString("EbpGBS_Service"),
		DisplayName: sec.Key("display_name").MustString("EbpGBS_Service"),
		Description: sec.Key("description").MustString("EbpGBS_Service"),
	}

	httpPort := mylog.Conf().Section("http").Key("port").MustInt(51180)
	oneHttpServer := NewOneHttpServer(httpPort)
	proxyPort := mylog.Conf().Section("proxy").Key("port").MustInt(7202)
	oneProxyServer := proxy.NewOneProxyServer(proxyPort)
	p := &application{}
	p.AddServer(oneHttpServer)
	p.AddServer(oneProxyServer)

	var s, err = service.New(p, svcConfig)
	if err != nil {
		log.Println(err)
		mylog.PauseExit()
	}
	if len(os.Args) > 1 {
		if os.Args[1] == "install" || os.Args[1] == "stop" {
			figure.NewFigure("Ebp-Proxy", "", false).Print()
		}
		log.Println(svcConfig.Name, os.Args[1], "...")
		if err = service.Control(s, os.Args[1]); err != nil {
			log.Println(err)
			mylog.PauseExit()
		}
		log.Println(svcConfig.Name, os.Args[1], "ok")
		return
	}
	figure.NewFigure("Ebp-Proxy", "", false).Print()
	if err = s.Run(); err != nil {
		log.Println(err)
		mylog.PauseExit()
	}
}
