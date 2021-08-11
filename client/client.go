package main
import (
    "fmt"
    "net"
    "bufio"
    "os"
    "encoding/json"
    "chatroom/common"
)

// Note:
// use the following command to see the output file watch -n 1 cat <username>.txt

func main() {

    fmt.Println("Chat client")
    fmt.Println("---------------------")

    username := read_username()

    for {
        msg := read_msg(username)

        if msg == "u" {
            // Update the output file with the new data from server
            req := build_request_get_updates(username)
            n, buf := sync_with_server(req)
            bulk_msg := common.UnmarshallBulkDbMessage(buf[:n])
            dump_msgs_to_output_file(username, bulk_msg)
        } else {
            // Send message
            req := build_request_send_message(username, msg)
            sync_with_server(req)
        }
    }
}

func dump_msgs_to_output_file(username string, bulk_msg common.BulkDbMessage) {
    fo, err := os.Create(username + ".txt")
    if err != nil {
        panic(err)
    }

    for i, v := range bulk_msg.DbMsgList {
        fo.Write([]byte(fmt.Sprintf("[%d] id: = %d, sender = %s, message = %s\n", i, v.Id, v.Sender, v.Msg)))
    }

    if err := fo.Close(); err != nil {
        panic(err)
    }
}

// read username from keyboard
func read_username() (username string) {
    reader := bufio.NewReader(os.Stdin)
    fmt.Println("Enter your username:")
    fmt.Print("-> ")
    s, _ := reader.ReadString('\n')
    username = s[:len(s)-1] // remove the new line
    return
}

// read message from keyboard
func read_msg(username string) (msg string) {
    reader := bufio.NewReader(os.Stdin)
    fmt.Printf("%v -> ", username)
    s, _ := reader.ReadString('\n')
    msg = s[:len(s)-1]
    return
}


// Create a serialized request to the server conatining the message
func build_request_send_message(username, msg string) (data []byte) {
    request := make(map[string]string)
    request["username"] = username
    request["function"] = "send_msg"
    request["msg"] = msg

    // var obj interface{}
    data, err := json.Marshal(request) 
    if err != nil {
        fmt.Println("Failed to marshall the request")
        fmt.Println(err)
    } else {
        fmt.Println("Request was marshalled")
        fmt.Println(string(data))
    }
    return
}

func build_request_get_updates(username string) (data []byte) {
    request := make(map[string]string)
    request["username"] = username
    request["function"] = "get_updates"

    // var obj interface{}
    data, err := json.Marshal(request) 
    if err != nil {
        // fmt.Println("Failed to marshall the request")
        fmt.Println(err)
    } else {
        fmt.Println("Request was marshalled")
        // fmt.Println(string(data))
    }
    return
}


func parse_response_from_server(n int, request []byte) (msg, username string) {
    var m map[string]string
    err := json.Unmarshal([]byte(request[:n]), &m)
    if err != nil {
        fmt.Println(err)
    }
    
    msg = m["msg"]
    username = m["username"]

    return
}


// Synchronize with server: send request and receive response
func sync_with_server(req []byte) (n int, buf []byte) {
    // Send the message to UDP server (chatroom)
    buf = make([]byte, 2048)
    conn, err := net.Dial("udp", "127.0.0.1:1234")
    if err != nil {
        fmt.Printf("Some error %v", err)
        return
    }
    // fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
    // fmt.Fprintf(conn, msg)
    fmt.Fprintf(conn, string(req))
    n, err = bufio.NewReader(conn).Read(buf)
    if err == nil {
        fmt.Printf("Server: %s\n", buf)
    } else {
        fmt.Printf("Server: ome error %v\n", err)
    }
    conn.Close()
    return
}