package routers

/**
 * @apiDefine sys API接口
 */

type APIHandler struct {
	RestartChan chan bool
}

var API = &APIHandler{
	RestartChan: make(chan bool),
}

func Init() (err error) {
	return nil
}
