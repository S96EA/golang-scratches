package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Logger interface {
	Errorf(format string, args... interface{})
	Infof(format string, args... interface{})
}

type stdLogger struct {
}

func newStdLogger() Logger{
	return &stdLogger{}
}

func (l *stdLogger) Errorf(format string, args... interface{}) {
	fmt.Printf("[ERROR] " + format + "\n", args...)
}

func (l *stdLogger) Infof(format string, args... interface{}) {
	fmt.Printf("[INFO]" + format + "\n", args...)
}

var logger = newStdLogger()

func main() {
	listen, err := net.Listen("tcp", "127.0.0.1:11223")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			panic(err)
		}
		go handle(conn)
	}
}

func parseStartLine(startLine string) (method, target, version string){
	split := strings.Split(startLine, " ")
	if len(split) != 3 {
		return "", "", ""
	}
	return split[0], split[1], split[2]
}

func handle(conn net.Conn) error {
	defer conn.Close()
	buffer := make([]byte, 1024*1024)

	var proxyConn net.Conn
	var err error

	_ = proxyConn
	_ = err

	for {
		n, err := conn.Read(buffer)
		if err != nil  && err != io.EOF{
			logger.Errorf("read error: %+v", err)
			return err
		}
		buffer[n] = 0
		//logger.Infof("read data: ", string(buffer))
		method, target, _ := parseStartLine(strings.Split(string(buffer), "\n")[0])
		switch method {
		case http.MethodConnect:
			proxyConn, err = processConnect(conn, target)
			if err != nil && err != io.EOF{
				return err
			}
			_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
			if err != nil {
				return err
			}
			go func() {
				io.Copy(proxyConn, conn)
			}()

			go func() {
				io.Copy(conn, proxyConn)
			}()
			time.Sleep(100*time.Second)
		}
	}
}

func processConnect(conn net.Conn, target string) (proxyConn net.Conn, err error){
	return net.Dial("tcp", target)
}
