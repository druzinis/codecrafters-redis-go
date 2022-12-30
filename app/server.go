package main

import (
	"fmt"
	 "net"
    "os"
    "io"
    "regexp"
    "strings"
)

type expirable struct {
    data []byte
    expires bool
    expiry int
}

// *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n
var regex = regexp.MustCompile(`^\*\d\r\n(\$\d+\r\n(.+)\r\n)(\$\d+\r\n(.+)\r\n)(\$\d+\r\n(.+)\r\n)?(\$\d+\r\n(.+)\r\n)?(\$\d+\r\n(.+)\r\n)?$`)
var m = make(map[string][]byte)

func main() {
	fmt.Println("Starting")
	 l, err := net.Listen("tcp", "0.0.0.0:6379")
	 if err != nil {
	 	fmt.Println("Failed to bind to port 6379")
	 	os.Exit(1)
     }
     for {
         conn, err := l.Accept()
         if err != nil {
             fmt.Println("Error accepting connection: ", err.Error())
             os.Exit(1)
         }
         go handleConn(conn)
     }
}

func handleConn(conn net.Conn) {
    var buffer = make([]byte, 100);
    for {
        var byteCount, err = conn.Read(buffer)
        if err != nil {
            if err == io.EOF {
                fmt.Println("Got EOF")
                return
            }
            fmt.Println("Error reading from connection: ", err.Error())
            os.Exit(1)
        }

        if buffer[0] != '*' || buffer[2] != '\r' {
            fmt.Println("Unexpected value:", buffer)
            return
        }
        if buffer[1] == '1' {
            // PING
            conn.Write([]byte("+PONG\r\n"))
        } else {
            var match = regex.FindSubmatch(buffer[:byteCount])
            var command = string(match[2])
            var first_arg = match[4];
            if strings.EqualFold("echo", command) {
                var send = append([]byte("+"), first_arg...)
                send = append(send, []byte("\r\n")...)
                conn.Write(send)
            } else if strings.EqualFold("set", command) {
                var second_arg = match[6]
                m[string(first_arg)] = second_arg
                sendPlainString(conn, []byte("OK"))
            } else if strings.EqualFold("get", command) {

                value, ok := m[string(first_arg)]
                if ok {
                    sendPlainString(conn, value)
                } else {
                    conn.Write([]byte("$-1\r\n"))
                }
            }
        }
//        var string = string(buffer[:byteCount])
//        var match = regex.FindSubmatch(buffer[:byteCount])
//        fmt.Println(buffer)
//        fmt.Println(string(buffer[:byteCount]))

//        if len(match) > 0 {
//
//        } else {
//
//            conn.Write([]byte("+PONG\r\n"))
//
//        }
    }
}

func sendPlainString(conn net.Conn, data []byte) {
    var send = append([]byte("+"), data...)
    send = append(send, []byte("\r\n")...)
    conn.Write(send)
}

//struct RedisCommand {
//    type RedisCommandType
//}
//
//type RedisCommandType int
//
//const (
//    PING RedisCommandType = []byte('PING')
//    ECHO = []byte('ECHO')
//)
