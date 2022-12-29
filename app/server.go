package main

import (
	"fmt"
	 "net"
    "os"
    "regexp"
)

// *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n
var regex = regexp.MustCompile(`^\*2\r\n\$\d+\r\n([A-Z]+)\r\n\$\d+\r\n(.*)\r\n$`)

func main() {
//    fmt.Println(make([]byte, 100)[:2])
//    fmt.Println("start")
//    var match = regexp.MustCompile(`^\*2\r\n\$(\d+)\r\n`).FindSubmatch([]byte("*2\r\n$456\r\n"))
//    fmt.Println(match[1])
//    fmt.Println("end")
//    fmt.Println("start")
//    match = regex.FindSubmatch([]byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"))
//    fmt.Println(match[1])
//    fmt.Println(match[2])
//
//    fmt.Println("end")
//    fmt.Println(regex.FindAllStringSubmatch("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n", 0))
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	 l, err := net.Listen("tcp", "0.0.0.0:6379")
	 if err != nil {
	 	fmt.Println("Failed to bind to port 6379")
	 	os.Exit(1)
     }
     for {
         var conn net.Conn
        conn, err = l.Accept()
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
            fmt.Println("Error reading from connection: ", err.Error())
            os.Exit(1)
        }
//        var string = string(buffer[:byteCount])
        var match = regex.FindSubmatch(buffer[:byteCount])
        if len(match) > 0 {

            var val = regex.FindSubmatch(buffer[:byteCount])[2]
            var send = append([]byte("+"), val...)
            send = append(send, []byte("\r\n")...)
            conn.Write(send)
        } else {

            conn.Write([]byte("+PONG\r\n"))
        }
    }
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
