package main

import (
    "fmt"
    "github.com/gomodule/redigo/redis"
    "net"
    "encoding/json"
    "chatroom/common"
)

const MSG_HISTORY_LEN = 5

// message id counter
var msg_id_ctr = 0

// database
var pool = newPool()

func main() {

    fmt.Println("Chat server")
    fmt.Println("---------------------")

    client := pool.Get()
    defer client.Close()

    _, err := client.Do("DEL", "messages")
    if err != nil {
        panic(err)
    }

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

        if err != nil {
            fmt.Println(err)
        }
        // fmt.Printf("Read a message from %v %s \n", remoteaddr, p)

        function, b := common.ParseRequest(p[:n])

        switch function {
        case common.F_SEND_MESSAGE:
            data := common.ParseSendMessageRequest(b)
            // fmt.Printf("data.Sender = %v\n", data.Sender)
            // fmt.Printf("data.Text = %v\n", data.Text)
            db_message := common.DbMessage{
                MsgId: msg_id_ctr,
                Sender: data.Sender,
                Text: data.Text,
            }
            msg_id_ctr++

            marshalled_msg, err := json.Marshal(db_message) 
            if err != nil {
                fmt.Println("Failed to marshall a message before writing to database")
                fmt.Println(err)
            }

            _, err = client.Do("RPUSH", "messages", marshalled_msg)
            if err != nil {
                panic(err)
            }

            // Limit history capacity
            _, err = client.Do("LTRIM", "messages", -MSG_HISTORY_LEN, -1)
            if err != nil {
                panic(err)
            }

            resp := common.BuildSendMessageResponse(true)

            go sendResponse(ser, remoteaddr, resp)
        case common.F_GET_UPDATES:

            messages, _ := redis.ByteSlices(client.Do("LRANGE", "messages", 0, -1))

            var resp_data common.GetUpdatesResponse
            for _, v := range messages {
                var m common.DbMessage
                err := json.Unmarshal(v, &m)
                if err != nil {
                    fmt.Println(err)
                }
                resp_data = append(resp_data, m)
            }

            // data := common.MarshallBulkDbMessage(&db_messages)

            resp := common.BuildGetUpdatesResponse(resp_data)
            go sendResponse(ser, remoteaddr, resp)

        default:
            fmt.Printf("Error, this should not happen\n")
        }
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

// server
func sendResponse(conn *net.UDPConn, addr *net.UDPAddr, buf []byte) {
    // _,err := conn.WriteToUDP([]byte("From server: Hello I got your message "), addr)
    _,err := conn.WriteToUDP(buf, addr)
    if err != nil {
        fmt.Printf("Couldn't send response %v", err)
    }
}
