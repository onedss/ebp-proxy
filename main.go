package main

import (
	"fmt"
	"github.com/onedss/ebp-proxy/app"
	"github.com/onedss/ebp-proxy/buildtime"
	"github.com/onedss/ebp-proxy/mylog"
	"github.com/onedss/ebp-proxy/utils"
	"log"
)

var (
	gitCommitCode string
	buildDateTime string
)

func main() {
	log.SetPrefix("[Ebp-Proxy] ")
	if mylog.Debug {
		log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	} else {
		log.SetFlags(log.LstdFlags)
	}
	buildtime.BuildVersion = fmt.Sprintf("%s.%s", buildtime.BuildVersion, gitCommitCode)
	buildtime.BuildTimeStr = fmt.Sprintf("<%s> %s", buildtime.BuildTime.Format(utils.DateTimeLayout), buildDateTime)
	mylog.Info("BuildVersion:", buildtime.BuildVersion)
	mylog.Info("BuildTime:", buildtime.BuildTimeStr)
	app.StartApp()
}
