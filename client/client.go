package main
import (
    "fmt"
    "net"
    "bufio"
    "os"
    // "encoding/json"
    "chatroom/common"
    "time"
    // "strings"
    // "strconv"
)

// Note:
// use the following command to see the output file watch -n 1 cat <username>.txt

func main() {

    fmt.Println("Chat client")
    fmt.Println("---------------------")

    username := read_username()
    go update_output_file(username)

    for {
        msg := read_msg(username)
        
        // if msg[0:1] == "d" {
            // fmt.Printf("delete msg, msg = [%v]\n", msg)
            // l := strings.Split(msg, " ")
            // msgid, err := strconv.Atoi(l[2])
            // if err == nil {
            //     req := common.BuildDeleteMessageRequest(username, msgid)
            //     sync_with_server(req)
            // }
            // b := common.ParseResponse(buf[:n])
            // data := common.ParseSendMessageResponse(b)
            // if data.Ok {
            //     fmt.Printf("Message with id %d was successfully deleted.\n")
            // } else {
            //     fmt.Printf("Error while deleting message with id %d:\n")
            //     fmt.Printf("%s\n", data.Error)
            // }
        // } else {
        // Send message
        req := common.BuildSendMessageRequest(username, msg)
        sync_with_server(req)
            
        // }
    }
}

func update_output_file(username string) {
    for {
        req := common.BuildGetUpdatesRequest()
        n, buf := sync_with_server(req)
        b := common.ParseResponse(buf[:n])
        data := common.ParseGetUpdatesResponse(b)
        dump_msgs_to_output_file(username, data)
        time.Sleep(1 * time.Second)
    }
}

func dump_msgs_to_output_file(username string, bulk_msg common.GetUpdatesResponse) {
    fo, err := os.Create(username + ".txt")
    if err != nil {
        panic(err)
    }

    for i, v := range bulk_msg {
        fo.Write([]byte(fmt.Sprintf("[%d] MsgId: = %d, sender = %s, text = %s\n", i, v.MsgId, v.Sender, v.Text)))
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
        // fmt.Printf("Server: %s\n", buf)
    } else {
        fmt.Printf("Server: ome error %v\n", err)
    }
    conn.Close()
    return
}
