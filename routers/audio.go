package routers

import (
	"github.com/gin-gonic/gin"
	ffmpeg "github.com/onedss/ebp-proxy/ffmpeg_go"
	"github.com/onedss/ebp-proxy/mytool"
	"log"
	"net/http"
)

type Convert2MP3Request struct {
	Source  string `form:"source" binding:"required"`
	Target  string `form:"target"`
	Bitrate string `form:"bitrate"`
}

func (h *APIHandler) Convert2MP3(c *gin.Context) {
	var err error
	var form Convert2MP3Request
	err = c.Bind(&form)
	if err != nil {
		log.Printf("Convert2MP3 Param Error: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if !mytool.Exist(form.Source) {
		log.Printf("Convert2MP3 Source Error: %s", form.Source)
		c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
		return
	}
	if form.Target == "" {
		form.Target = form.Source + ".mp3"
	}
	bitrate := mytool.Conf().Section("ffmpeg").Key("audio_bitrate").MustString("64k")
	if form.Bitrate != "" {
		bitrate = form.Bitrate
	}
	log.Println("Convert2MP3 >>> Source:", form.Source, "; Target:", form.Target, "; Bitrate:", bitrate)
	err = ffmpeg.Input(form.Source).
		Output(form.Target, ffmpeg.KwArgs{"b:a": bitrate}).
		OverWriteOutput().ErrorToStdOut().Run()
	if err != nil {
		c.JSON(http.StatusBadRequest, "ERROR")
	} else {
		c.JSON(http.StatusOK, "OK")
	}
}
