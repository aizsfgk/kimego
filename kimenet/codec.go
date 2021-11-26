package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type GraceInfo struct {
	Fd int
	State int
	ReadBuff []byte
	WriteBuff []byte
	idx string
}

func Encode(conn *Connection) (bt []byte, err error){

	gi := GraceInfo{}
	gi.Fd = conn.Fd
	gi.State = conn.State
	gi.ReadBuff = conn.ReadBuff.Bytes()
	gi.WriteBuff = conn.ReadBuff.Bytes()
	gi.idx = conn.idx

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err = enc.Encode(gi); err != nil {
		return
	}

	bt = buf.Bytes()
	return
}

func Decode(bt []byte) (conn *Connection, err error) {
	var c GraceInfo
	var buf bytes.Buffer
	_, err = buf.Write(bt)
	if err != nil {
		fmt.Println("buf.Write err: ", err.Error())
		return
	}

	dec := gob.NewDecoder(&buf)
	if err = dec.Decode(&c); err != nil {
		fmt.Println("dec.Decode err: ", err.Error())
		return
	}

	conn = &Connection{
		Fd:        c.Fd,
		State:     c.State,
		ReadBuff:  bytes.NewBuffer(c.ReadBuff),
		WriteBuff: bytes.NewBuffer(c.WriteBuff),
		idx:       c.idx,
	}
	return
}
