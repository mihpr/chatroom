package main

import (
    "fmt"
    "github.com/gomodule/redigo/redis"
    "net"
    // "encoding/json"
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

            msg_id := get_next_msg_id(client)

            // Add message IDs and message data
            _, err = client.Do("RPUSH", "msg_ids", msg_id)
            if err != nil {
                panic(err)
            }

            hash_key := get_hash_key_by_msg_id(msg_id)

            _, err = client.Do("HSET", hash_key, "MsgId", msg_id)
            if err != nil {
                panic(err)
            }

            _, err = client.Do("HSET", hash_key, "Sender", data.Sender)
            if err != nil {
                panic(err)
            }

            _, err = client.Do("HSET", hash_key, "Text", data.Text)
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
                hash_key := get_hash_key_by_msg_id(msg_id)
                MsgId, err := redis.Int64(client.Do("HGET", hash_key, "MsgId"))
                if err != nil {
                    panic(err)
                }
                Sender, err := redis.String(client.Do("HGET", hash_key, "Sender"))
                if err != nil {
                    panic(err)
                }
                Text, err := redis.String(client.Do("HGET", hash_key, "Text"))
                if err != nil {
                    panic(err)
                }
                message := common.MessageData {
                    MsgId: MsgId,
                    Sender: Sender,
                    Text: Text,
                }

                resp_data = append(resp_data, message)
            }

            resp := common.BuildGetUpdatesResponse(resp_data)
            go sendResponse(ser, remoteaddr, resp)

        case common.F_DEL_MESSAGE:
            req_data := common.ParseDeleteMessageRequest(b)

            hash_key := get_hash_key_by_msg_id(req_data.MsgId)

            key_exists, err := redis.Int(client.Do("EXISTS", hash_key))
            if err != nil {
                panic(err)
            }

            var resp_data common.DeleteMessageResponse
            if key_exists == 0 {
                resp_data.Ok = false
                resp_data.Error = fmt.Sprintf("There is no message with ID [%v]", req_data.MsgId)
            } else {
                Sender, err := redis.String(client.Do("HGET", hash_key, "Sender"))
                if err != nil {
                    panic(err)
                }
                if Sender == req_data.Sender {
                    count, err := redis.Int(client.Do("LREM", "msg_ids", 0, req_data.MsgId))
                    if err != nil {
                        panic(err)
                    }
                    _, err = client.Do("DEL", hash_key)
                    if count == 1 {
                        resp_data.Ok = true
                        // resp_data.Error = fmt.Sprintf("Message with ID [%v]", req_data.MsgId)
                        resp_data.Error = ""
                    } else {
                        resp_data.Ok = false
                        resp_data.Error = fmt.Sprintf("Unexpected error while delering message with ID [%v]", req_data.MsgId)
                    }
                } else {
                    resp_data.Ok = false
                    resp_data.Error = fmt.Sprintf("It is forbidden to delete message with ID [%v], since it is not yours", req_data.MsgId)
                }
            }
            
            

            resp := common.BuildDeleteMessageResponse(resp_data)
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
