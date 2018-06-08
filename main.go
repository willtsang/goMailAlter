package main

import (
	"net/http"
	"fmt"
	"runtime"
	"time"
	"strings"
	"net/smtp"
	_ "net/http/pprof"
	log "github.com/thinkboy/log4go"
)

var (
	Logger  log.Logger
	MailErr = make(chan error, 4)
)

func main() {

	defer Logger.Close()

	urls := []string{
		"http://www.sina.com",
		"http://www.baidu.com",
		"http://www.jd.com",
		"http://www.qq.com",
		"http://www.163.com",
		"http://www.tmall.com",
	}

	for _, url := range urls {
		go isAvailable(url)
	}

	var exit chan bool

	<-exit
}

// Init func
func init() {

	setLogger()
	setMonitor()

	mailHandel()
}

// Handle mail goroutine
func mailHandel() {

	go func() {
		for {
			mailErr := <-MailErr

			alterMail(mailErr)
		}
	}()
}

// Set Default Logger
func setLogger() {

	Logger = make(log.Logger)

	Logger.AddFilter("stdout", log.ERROR, log.NewConsoleLogWriter())
	Logger.AddFilter("log", log.DEBUG, log.NewFileLogWriter("log.log", false))
}

// Set Monitor
func setMonitor() {
	go func() {
		http.ListenAndServe("0.0.0.0:8808", nil)
	}()
}

// Is the Server Available
func isAvailable(url string) {

	httpClient := &http.Client{
		Timeout: time.Duration(1) * time.Second,
	}

	for {
		response, err := httpClient.Get(url)
		var m runtime.MemStats

		Logger.Debug("request:%s", url)

		if err != nil {
			Logger.Error("request %v error", err.Error())
			MailErr <- err
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}

		statusCode := response.StatusCode

		if statusCode != 200 {
			Logger.Warn("request response is not 200, status code is %d", statusCode)
			MailErr <- err
		}

		response.Body.Close()

		runtime.ReadMemStats(&m)

		time.Sleep(time.Duration(10) * time.Second)

		Logger.Debug("heapAlloc:%dM", m.HeapAlloc/1024/1024)
	}
}

// Send alter Mail
func alterMail(serverErr error) {

	user := "username@mail.com"
	password := "password_stirng"
	host := "smtp.exmail.qq.com:25"
	to := "example@mail.com"

	subject := "Service is not available"
	body := fmt.Sprintf(" <html> <body> <h3> Server is not available : %s </h3> </body> </html>", serverErr.Error())

	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])

	contentType := "Content-Type: text/html; charset=UTF-8"

	msg := []byte("To: " + to + "\r\nFrom: " + user + "\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")

	Logger.Debug("Send Mail")

	err := smtp.SendMail(host, auth, user, sendTo, msg)

	if err != nil {
		Logger.Error("Send mail error! info:%v", err.Error())
	}
}
