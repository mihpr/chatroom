module server

go 1.16

require (
	chatroom/common v0.0.0-00010101000000-000000000000
	github.com/gomodule/redigo v1.8.5
)

replace chatroom/common => ../common
