package common

import (
    "fmt"
    "encoding/json"
)


// ----------------------------------------------------------------------------------------------------
// Database
// ----------------------------------------------------------------------------------------------------

// ----------------------------------------------------------------------------------------------------
// Request and response
// ----------------------------------------------------------------------------------------------------

const F_SEND_MESSAGE = "send_msg"
const F_GET_UPDATES  = "get_updates"
const F_DEL_MESSAGE  = "del_msg"

type MessageData struct {
    MsgId int64
    Sender string
    Text string
}

type Request struct {
    Function string
    Data []byte
}

type Response struct {
    // Function string
    Data []byte
}

func BuildRequest (function string, data []byte) (b []byte) {
    req := Request {
        Function: function,
        Data: data,
    }

    b, err := json.Marshal(req)
    if err != nil {
        fmt.Println("Error in BuildRequest() function")
        fmt.Println(err)
    }
    return
}

func ParseRequest (req []byte) (function string, data []byte) {
    var request Request
    err := json.Unmarshal(req, &request)
    if err != nil {
        fmt.Println("Error in ParseRequest() function")
        fmt.Println(err)
    }
    return request.Function, request.Data
}


func BuildResponse(data []byte) (b []byte) {
    resp := Response {
        Data: data,
    }

    b, err := json.Marshal(resp)
    if err != nil {
        fmt.Println("Error in BuildResponse() function")
        fmt.Println(err)
    }
    return
}

func ParseResponse (resp []byte) (data []byte) {
    var response Response
    err := json.Unmarshal(resp, &response)
    if err != nil {
        fmt.Println("Error in ParseResponse() function")
        fmt.Println(err)
    }
    return response.Data
}

// ----------------------------------------------------------------------------------------------------
// Function send message
// ----------------------------------------------------------------------------------------------------

type SendMessageRequest struct {
    Sender string
    Text string
}

type SendMessageResponse bool // true if there are no errors

func BuildSendMessageRequest(sender string, text string) (b []byte) {
    data := SendMessageRequest {
        Sender: sender,
        Text: text,
    }

    b, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error in BuildSendMessageRequest() while marshaling data")
        fmt.Println(err)
    }

    b = BuildRequest(F_SEND_MESSAGE, b)
    return
}

func ParseSendMessageRequest(b []byte) (data SendMessageRequest) {
    err := json.Unmarshal(b, &data)
    if err != nil {
        fmt.Println("Error in ParseSendMessageRequest() while unmarshaling data")
        fmt.Println(err)
    }
    return
}

func BuildSendMessageResponse(ok bool) (b []byte) {
    data := SendMessageResponse(true)

    b, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error in BuildSendMessageResponse() while marshaling data")
        fmt.Println(err)
    }

    b = BuildResponse(b)
    return
}

func ParseSendMessageResponse(b []byte) (data SendMessageResponse) {
    err := json.Unmarshal(b, &data)
    if err != nil {
        fmt.Println("Error in ParseendMessageResponse() while unmarshaling data")
        fmt.Println(err)
    }
    return
}

// ----------------------------------------------------------------------------------------------------
// Function get updates
// ----------------------------------------------------------------------------------------------------

type GetUpdatesRequest struct {
}

type GetUpdatesResponse []MessageData

func BuildGetUpdatesRequest() (b []byte) {
    b, err := json.Marshal(nil)
    if err != nil {
        fmt.Println("Error in BuildGetUpdatesRequest() while marshaling data")
        fmt.Println(err)
    }

    b = BuildRequest(F_GET_UPDATES, b)
    return
}

// func ParseGetUpdatesRequest() is not required

func BuildGetUpdatesResponse(dbMsgList []MessageData) (b []byte) {
    // data := dbMsgList
    b, err := json.Marshal(dbMsgList)
    if err != nil {
        fmt.Println("Error in BuildGetUpdatesRequest() while marshaling data")
        fmt.Println(err)
    }

    b = BuildResponse(b)
    return
}

func ParseGetUpdatesResponse(b []byte) (data GetUpdatesResponse) {
    err := json.Unmarshal(b, &data)
    if err != nil {
        fmt.Println("Error in ParseendMessageResponse() while unmarshaling data")
        fmt.Println(err)
    }
    return
}

// ----------------------------------------------------------------------------------------------------
// Function delete message
// ----------------------------------------------------------------------------------------------------
type DeleteMessageRequest struct {
    Sender string
    MsgId int
}

type DeleteMessageResponse struct {
    Ok bool
    Error string
}

func BuildDeleteMessageRequest(sender string, msgid int) (b []byte) {
    data := DeleteMessageRequest {
        Sender: sender,
        MsgId: msgid,
    }

    b, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error in BuildDeleteMessageRequest() while marshaling data")
        fmt.Println(err)
    }

    b = BuildRequest(F_DEL_MESSAGE, b)
    return
}

func ParseDeleteMessageRequest(b []byte) (data DeleteMessageRequest) {
    err := json.Unmarshal(b, &data)
    if err != nil {
        fmt.Println("Error in ParseDeleteMessageRequest() while unmarshaling data")
        fmt.Println(err)
    }
    return
}

func BuildDeleteMessageResponse(ok bool, error string) (b []byte) {
    data := DeleteMessageResponse {
        Ok: ok,
        Error: error,
    }
    b, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error in BuildDeleteMessageResponse() while marshaling data")
        fmt.Println(err)
    }

    b = BuildResponse(b)
    return
}

func ParseDeleteMessageResponse(b []byte) (data DeleteMessageResponse) {
    err := json.Unmarshal(b, &data)
    if err != nil {
        fmt.Println("Error in ParseDeleteMessageResponse() while unmarshaling data")
        fmt.Println(err)
    }
    return
}