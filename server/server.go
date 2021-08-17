package main

import (
    "fmt"
    "github.com/gomodule/redigo/redis"
    "net"
    "encoding/json"
    "chatroom/common"
)

const MSG_HISTORY_LEN = 5

// database
var pool = newPool()

func main() {

    fmt.Println("Chat server")
    fmt.Println("---------------------")

    client := pool.Get()
    defer client.Close()
    db_clenup(client)

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
            message := common.Message{
                MsgId: get_next_msg_id(client),
                Sender: data.Sender,
                Text: data.Text,
            }

            marshalled_msg, err := json.Marshal(message) 
            if err != nil {
                fmt.Println("Failed to marshall a message before writing to database")
                fmt.Println(err)
            }

            // Add message IDs and message data
            _, err = client.Do("RPUSH", "msg_ids", message.MsgId)
            if err != nil {
                panic(err)
            }

            hash_key := get_hash_key_by_msg_id(message.MsgId)
            _, err = client.Do("HSET", hash_key, "data", marshalled_msg) // TODO: replace marshalling by HMSETing the data items separately (?)
            if err != nil {
                panic(err)
            }

            // Limit history capacity
            int64s, err := redis.Int64s(client.Do("LRANGE", "msg_ids", 0, -MSG_HISTORY_LEN - 1))
            if err != nil {
                panic(err)
            }

            for _, msg_id := range(int64s) {
                hash_key := get_hash_key_by_msg_id(msg_id)
                _, err = client.Do("DEL", hash_key)
                if err != nil {
                    panic(err)
                }
            }

            _, err = client.Do("LTRIM", "msg_ids", -MSG_HISTORY_LEN, -1)
            if err != nil {
                panic(err)
            }

            resp := common.BuildSendMessageResponse(true)

            go sendResponse(ser, remoteaddr, resp)
        case common.F_GET_UPDATES:

            int64s, err := redis.Int64s(client.Do("LRANGE", "msg_ids", 0, -1))
            if err != nil {
                panic(err)
            }

            var resp_data common.GetUpdatesResponse
            for _, msg_id := range(int64s) {
                // fmt.Printf("msg_id = [%d]\n", msg_id)
                // fmt.Printf("msg_id format [%T]\n", msg_id)
                hash_key := fmt.Sprintf("msg_data:%d", msg_id)
                marshalled_msg, err := redis.Bytes(client.Do("HGET", hash_key, "data"))
                if err != nil {
                    panic(err)
                }

                var m common.Message
                err = json.Unmarshal(marshalled_msg, &m)
                if err != nil {
                    fmt.Println(err)
                }
                resp_data = append(resp_data, m)
            }

            resp := common.BuildGetUpdatesResponse(resp_data)
            go sendResponse(ser, remoteaddr, resp)

        // case common.F_DEL_MESSAGE:
            // req_data := common.ParseDeleteMessageRequest(b)


            // messages, _ := redis.ByteSlices(client.Do("LRANGE", "messages", 0, -1))

            // // var resp_data common.GetUpdatesResponse
            // for idx, v := range messages {
            //     var m common.Message
            //     err := json.Unmarshal(v, &m)
            //     if err != nil {
            //         fmt.Println(err)
            //     }
            //     // resp_data = append(resp_data, m)
            //     if m.MsgId == req_data.MsgId {
            //         if m.Sender == req_data.Sender {
            //             // TODO: delete message

            //         } else {
            //             // TODO: this is not yours message
            //         }

            //     } else {
            //         // TODO: message not found
            //     }

            // }


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

func get_next_msg_id(client redis.Conn) (int64) {
    next_user_id, err := client.Do("INCR", "msgid_counter")
    if err != nil {
        panic(err)
    }
    ret, ok := next_user_id.(int64)
    if ok {
        return ret
    }
    return 0
}

func db_clenup(client redis.Conn) {
    keys, err := redis.Strings(client.Do("KEYS", "msg_data:*"))
    if err != nil {
        panic(err)
    }

    for _, key := range keys {
        _, err = client.Do("DEL", key)
        if err != nil {
            panic(err)
        }
        // fmt.Printf("deleting key = [%s]\n", key)
    }


    _, err = client.Do("DEL", "msg_ids")
    if err != nil {
        panic(err)
    }

    _, err = client.Do("SET", "msgid_counter", 0)
    if err != nil {
        panic(err)
    }
}

func get_hash_key_by_msg_id(msg_id int64) (hash_key string) {
    hash_key = fmt.Sprintf("msg_data:%d", msg_id)
    return
}

// server
func sendResponse(conn *net.UDPConn, addr *net.UDPAddr, buf []byte) {
    // _,err := conn.WriteToUDP([]byte("From server: Hello I got your message "), addr)
    _,err := conn.WriteToUDP(buf, addr)
    if err != nil {
        fmt.Printf("Couldn't send response %v", err)
    }
}
