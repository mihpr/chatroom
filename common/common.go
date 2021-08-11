package common

import (
    "fmt"
    "encoding/json"
)

type DbMessage struct {
    Id int
    Sender string
    Msg string
}

type BulkDbMessage struct {
    DbMsgList []DbMessage
}


func MarshallBulkDbMessage(msg *BulkDbMessage) ([]byte) {
    marshalled, err := json.Marshal(msg)
    if err != nil {
        fmt.Println(err)
    }
    return marshalled
}

func UnmarshallBulkDbMessage(buf []byte) (buld_db_message BulkDbMessage) {
    err := json.Unmarshal(buf, &buld_db_message)
    if err != nil {
        fmt.Println(err)
    }
    return
}