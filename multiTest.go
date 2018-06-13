package main

import (
	log "github.com/thinkboy/log4go"

	"net/http"
)

var (
	Logger log.Logger
)

func main() {

	defer Logger.Close()

	begin := 0

	for i := begin; i < 100; i++ {
		go test(i)
	}

	var exit chan bool

	<-exit
}

func init() {
	setLogger()
}

func test(gorountineNum int) {

	for {
		log.Debug("start%d", gorountineNum)

		response, err := http.Get("http://www.baidu.com")

		if err != nil {
			panic(err.Error())
		}

		response.Body.Close()
	}
}

// Set Default Logger
func setLogger() {

	Logger = make(log.Logger)

	Logger.AddFilter("stdout", log.DEBUG, log.NewConsoleLogWriter())
	Logger.AddFilter("log", log.DEBUG, log.NewFileLogWriter("log.log", false))
}
