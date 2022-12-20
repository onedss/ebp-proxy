package app

import (
	"context"
	"encoding/json"
	"fmt"
	ffmpeg "github.com/onedss/ebp-proxy/ffmpeg_go"
	"github.com/onedss/ebp-proxy/mytool"
	"github.com/onedss/ebp-proxy/routers"
	"io"
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
		Addr:              fmt.Sprintf(":%d", p.httpPort),
		Handler:           routers.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	//http.HandleFunc("/", homeHandler)
	//http.HandleFunc("/status", myHandler)
	//http.HandleFunc("/ffmpeg", ffmpegHandler)
	link := fmt.Sprintf("http://%s:%d", mytool.LocalIP(), p.httpPort)
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
	log.Println(r.RemoteAddr, "连接成功")
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
	log.Println(r.RemoteAddr, "连接成功")
	// 请求方式：GET POST DELETE PUT UPDATE
	fmt.Println("method:", r.Method)
	// /go
	fmt.Println("url:", r.URL.Path)
	fmt.Println("header:", r.Header)
	fmt.Println("body:", r.Body)
	// 回复
	w.Write([]byte("OK"))
}

func ffmpegHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, "收到请求成功")
	runExampleStream("in1.mp4", "out1.mp4")
	// 回复
	w.Write([]byte("OK"))
	log.Println(r.RemoteAddr, "处理请求成功")
}

func runExampleStream(inFile, outFile string) {
	w, h := getVideoSize(inFile)
	log.Println(w, h)

	pr1, pw1 := io.Pipe()
	pr2, pw2 := io.Pipe()
	done1 := startFFmpegProcess1(inFile, pw1)
	process(pr1, pw2, w, h)
	done2 := startFFmpegProcess2(outFile, pr2, w, h)
	err := <-done1
	if err != nil {
		panic(err)
	}
	err = <-done2
	if err != nil {
		panic(err)
	}
	log.Println("Done")
}

func getVideoSize(fileName string) (int, int) {
	log.Println("Getting video size for", fileName)
	data, err := ffmpeg.Probe(fileName)
	if err != nil {
		panic(err)
	}
	log.Println("got video info", data)
	type VideoInfo struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int
			Height    int
		} `json:"streams"`
	}
	vInfo := &VideoInfo{}
	err = json.Unmarshal([]byte(data), vInfo)
	if err != nil {
		panic(err)
	}
	for _, s := range vInfo.Streams {
		if s.CodecType == "video" {
			return s.Width, s.Height
		}
	}
	return 0, 0
}

func startFFmpegProcess1(infileName string, writer io.WriteCloser) <-chan error {
	log.Println("Starting ffmpeg process1")
	done := make(chan error)
	go func() {
		err := ffmpeg.Input(infileName).
			Output("pipe:",
				ffmpeg.KwArgs{
					"format": "rawvideo", "pix_fmt": "rgb24",
				}).
			WithOutput(writer).
			Run()
		log.Println("ffmpeg process1 done")
		_ = writer.Close()
		done <- err
		close(done)
	}()
	return done
}

func startFFmpegProcess2(outfileName string, buf io.Reader, width, height int) <-chan error {
	log.Println("Starting ffmpeg process2")
	done := make(chan error)
	go func() {
		err := ffmpeg.Input("pipe:",
			ffmpeg.KwArgs{"format": "rawvideo",
				"pix_fmt": "rgb24", "s": fmt.Sprintf("%dx%d", width, height),
			}).
			Output(outfileName, ffmpeg.KwArgs{"pix_fmt": "yuv420p"}).
			OverWriteOutput().
			WithInput(buf).
			Run()
		log.Println("ffmpeg process2 done")
		done <- err
		close(done)
	}()
	return done
}

func process(reader io.ReadCloser, writer io.WriteCloser, w, h int) {
	go func() {
		frameSize := w * h * 3
		buf := make([]byte, frameSize, frameSize)
		for {
			n, err := io.ReadFull(reader, buf)
			if n == 0 || err == io.EOF {
				_ = writer.Close()
				return
			} else if n != frameSize || err != nil {
				panic(fmt.Sprintf("read error: %d, %s", n, err))
			}
			for i := range buf {
				buf[i] = buf[i] / 3
			}
			n, err = writer.Write(buf)
			if n != frameSize || err != nil {
				panic(fmt.Sprintf("write error: %d, %s", n, err))
			}
		}
	}()
	return
}
