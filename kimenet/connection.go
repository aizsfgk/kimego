package main

import "bytes"


const (
	ESTABLISHED = 1
	CLOSED = 0
)

type Connection struct {
	Fd int
	State int
	ReadBuff *bytes.Buffer
	WriteBuff *bytes.Buffer
	idx string
}


func NewConnection(fd int, idx string) (conn *Connection, err error) {
	conn = new(Connection)
	conn.Fd = fd
	conn.idx = idx
	conn.ReadBuff = bytes.NewBuffer(make([]byte, 1024))
	conn.WriteBuff = bytes.NewBuffer(make([]byte, 1024))
	conn.State = ESTABLISHED
	return
}
