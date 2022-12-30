package main

import (
	"fmt"
	 "net"
    "os"
    "io"
    "regexp"
    "strings"
    "hash/maphash"
    "sync"
    "strconv"
    "time"
)

type expirable struct {
    data *[]byte
    expires bool
    expiresAt int64
}

func createNonexpirable(value *[]byte) *expirable {
    return &expirable{data: value, expires: false, expiresAt: 0}
}

func createExpirable(value *[]byte, expiry int64) *expirable {
    expiresAt := time.Now().UnixMilli() + expiry
    return &expirable{data: value, expires: true, expiresAt: expiresAt}
}

// *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n
var regex = regexp.MustCompile(`^\*\d\r\n(\$\d+\r\n(.+)\r\n)(\$\d+\r\n(.+)\r\n)(\$\d+\r\n(.+)\r\n)?(\$\d+\r\n(.+)\r\n)?(\$\d+\r\n(.+)\r\n)?$`)
var m = make(map[string]*expirable)
var locks [32]*sync.Mutex

func main() {
    for i, _ := range locks {
        var m sync.Mutex
        locks[i] = &m
    }

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
                var value *expirable
                if (len(match[10]) > 0) {
                    // with expiry
                    expiry, err := strconv.ParseInt(string(match[10]), 10, 64)
                    if err != nil {
                        fmt.Println("Error parsing expiry", match[10], string(match[10]))
                        value = createNonexpirable(&second_arg)
                    } else {
                        value = createExpirable(&second_arg, expiry)
                    }
                } else {
                    value = createNonexpirable(&second_arg)
                }
                lock := getLock(&first_arg)
                lock.Lock()
                m[string(first_arg)] = value
                lock.Unlock()
                conn.Write([]byte("+OK\r\n"))
            } else if strings.EqualFold("get", command) {

                lock := getLock(&first_arg)
                lock.Lock()
                value, ok := m[string(first_arg)]
                if ok && value.expires && time.Now().UnixMilli() > value.expiresAt {
                    delete(m, string(first_arg))
                    value = nil
                }
                lock.Unlock()
                if value != nil {
                    sendPlainString(conn, value.data)
                } else {
                    conn.Write([]byte("$-1\r\n"))
                }
            }
        }
    }
}

func getLock(value *[]byte) *sync.Mutex {
    var h maphash.Hash
    h.Write(*value)
    return locks[h.Sum64() % 32]
}

func sendPlainString(conn net.Conn, data *[]byte) {
    var send = append([]byte("+"), *data...)
    send = append(send, []byte("\r\n")...)
    conn.Write(send)
}
