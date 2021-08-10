package main

import (
    "fmt"
    "github.com/gomodule/redigo/redis"
    "net"
    "encoding/json"

)


// database
var pool = newPool()

func main() {

    // client := pool.Get()
    // defer client.Close()

    // _, err := client.Do("SET", "mykey", "Hello from redigo!")
    // if err != nil {
    //     panic(err)
    // }

    // value, err := client.Do("GET", "mykey")
    // if err != nil {
    //     panic(err)
    // }

    // fmt.Printf("%s \n", value)

    // _, err = client.Do("ZADD", "vehicles", 4, "car")
    // if err != nil {
    //     panic(err)
    // }
    // _, err = client.Do("ZADD", "vehicles", 2, "bike")
    // if err != nil {
    //     panic(err)
    // }

    // vehicles, err := client.Do("ZRANGE", "vehicles", 0, -1, "WITHSCORES")
    // if err != nil {
    //     panic(err)
    // }  
    // fmt.Printf("%s \n", vehicles)


    // UDP server
    p := make([]byte, 2048)
    addr := net.UDPAddr{
        Port: 1234,
        IP: net.ParseIP("127.0.0.1"),
    }
    ser, err := net.ListenUDP("udp", &addr)
    if err != nil {
        fmt.Printf("Some error %v\n", err)
        return
    }
    for {
        n,remoteaddr,err := ser.ReadFromUDP(p)
        fmt.Printf("Read a message from %v %s \n", remoteaddr, p)

        fmt.Printf("--- \n")
        msg, username := parse_request_from_client(n, p)
        fmt.Printf("msg = [%v]\n", msg)
        fmt.Printf("username = [%v]\n", username)
        fmt.Printf("--- \n")

        if err !=  nil {
            fmt.Printf("Some error  %v", err)
            continue
        }
        go sendResponse(ser, remoteaddr)
    }
}


// database
func newPool() *redis.Pool {
    return &redis.Pool{
        MaxIdle: 80,
        MaxActive: 12000,
        Dial: func() (redis.Conn, error) {
            c, err := redis.Dial("tcp", ":6379")
            if err != nil {
                panic(err.Error())
            }
            return c, err
        },
    }
}

// udp server
func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
    _,err := conn.WriteToUDP([]byte("From server: Hello I got your message "), addr)
    if err != nil {
        fmt.Printf("Couldn't send response %v", err)
    }
}

func parse_request_from_client(n int, request []byte) (msg, username string) {
    var m map[string]string
    err := json.Unmarshal([]byte(request[:n]), &m)
    if err != nil {
        fmt.Println(err)
    }
    
    msg = m["msg"]
    username = m["username"]

    return
}

func build_response_to_client(username, msg string) (data []byte) {
    response := make(map[string]string)
    response["username"] = username
    response["msg"] = msg

    // var obj interface{}
    data, err := json.Marshal(response) 
    if err != nil {
        fmt.Println("Failed to marshall the response")
        fmt.Println(err)
    }
    // else {
    //     fmt.Println("Request was marshalled")
    //     fmt.Println(string(data))
    // }
    return
}