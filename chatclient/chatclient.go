package main
import (
    "fmt"
    "net"
    "bufio"
    // "strings"
    "os"
    "encoding/json"
)

func main() {

    username := read_username()

    for {
        msg := read_msg(username)
        request := build_request(username, msg)
        poll_server(request)
    }
}

// read username from keyboard
func read_username() (username string) {
    reader := bufio.NewReader(os.Stdin)
    fmt.Println("Simple Shell")
    fmt.Println("---------------------")

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
   
    // if strings.Compare("q", msg) == 0 {
    //     fmt.Println("Exiting...")
    //     break
    // }
    return
}


// Create a serialized request to the server conatining the message
func build_request(username, msg string) (data []byte) {
    request := make(map[string]string)
    request["username"] = username
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


// Poll server for updates
func poll_server(request []byte) {
    // Send the message to UDP server (chatroom)
    p :=  make([]byte, 2048)
    conn, err := net.Dial("udp", "127.0.0.1:1234")
    if err != nil {
        fmt.Printf("Some error %v", err)
        return
    }
    // fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
    // fmt.Fprintf(conn, msg)
    fmt.Fprintf(conn, string(request))
    _, err = bufio.NewReader(conn).Read(p)
    if err == nil {
        fmt.Printf("Server: %s\n", p)
    } else {
        fmt.Printf("Server: ome error %v\n", err)
    }
    conn.Close()
    return
}