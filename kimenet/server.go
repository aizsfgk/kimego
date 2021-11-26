package main

import (
	"fmt"
	"net"
	"syscall"
)

const (
	Start   = 0
	Stop    = 1
	Gracing = 2
)

type Server struct {
	State        int
	Ip           string
	Port         int
	ListenFd     int
	UnixListener *net.UnixListener
	UnixServer   *net.UnixConn
	evloop       *EventLoop
	Conns        map[int]*Connection
}

func NewServer(ip string, port int) (srv *Server, err error) {
	srv = new(Server)
	// 1. create socketfd
	socketFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, 0)
	if err != nil || socketFd < 0 {
		err = fmt.Errorf("syscall.Socket ERR")
		return
	}

	// 2. bind addr and port
	ip4 := net.ParseIP(ip).To4()
	if ip4 == nil {
		err = fmt.Errorf("net.ParseIP err")
		return
	}
	sa := &syscall.SockaddrInet4{Port: port}
	copy(sa.Addr[:], ip4)

	err = syscall.SetsockoptInt(socketFd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return
	}
	err = syscall.SetsockoptInt(socketFd, syscall.SOL_SOCKET, 0xf, 1)
	if err != nil {
		return
	}

	err = syscall.Bind(socketFd, sa)
	if err != nil {
		err = fmt.Errorf("socket bind err: %s", err.Error())
		return
	}

	// 3. listen
	err = syscall.Listen(socketFd, 128)
	if err != nil {
		err = fmt.Errorf("socket listen err: %s", err)
		return
	}

	srv.Ip = ip
	srv.Port = port
	srv.ListenFd = socketFd
	srv.Conns = make(map[int]*Connection, 1<<5)

	return
}

func (srv *Server) EventLoopCallback(fd int, event Event) (err error) {

	fmt.Println("fd: ", fd)
	fmt.Println("ListenFd: ", srv.ListenFd)

	if fd == srv.ListenFd {
		return srv.HandleAccept(fd)
	}

	if (event & READ_EVENT) != 0 {
		err = srv.HandleRead(fd)
	}

	if (event & WRITE_EVNET) != 0 {
		err = srv.HandleWrite(fd)
	}

	return
}

func (srv *Server) HandleAccept(fd int) (err error) {
	var (
		acceptedFd int
		sa         syscall.Sockaddr
		conn       *Connection
	)
	fmt.Println("开始接收FD")
	acceptedFd, sa, err = syscall.Accept4(fd, syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC) // syscall.SOCK_CLOEXEC 这里不能有这个
	if err != nil {
		if err == syscall.EAGAIN {
			return
		}

		fmt.Println("accept err: ", err.Error())
		return
	}

	conn, err = NewConnection(acceptedFd, sockAddrToString(sa))
	if err != nil {
		return
	}
	fmt.Println("新建连接成功")
	srv.Conns[acceptedFd] = conn
	err = srv.evloop.AddRead(acceptedFd)

	return
}

func (srv *Server) HandleRead(fd int) (err error) {
	var (
		buf       = make([]byte, 1024)
		readN     int
		bufWriteN int
	)

	if conn, ok := srv.Conns[fd]; ok {
		readN, err = syscall.Read(conn.Fd, buf)
		if err != nil {
			if err == syscall.EAGAIN {
				err = nil
				fmt.Println("READ EAGIN")
				return
			}
			fmt.Println("Read err: ", err.Error())
			_ = srv.CloseFd(fd)
			return
		}

		if readN == 0 {
			fmt.Println("Read EOL")
			_ = srv.CloseFd(fd)
			return
		}

		if readN > 0 {
			conn.ReadBuff.Write(buf[:readN])
		}

		// 业务逻辑处理
		if conn.ReadBuff.Len() > 0 {
			bufWriteN, _ = conn.WriteBuff.Write(conn.ReadBuff.Bytes())
			if bufWriteN > 0 {
				_ = srv.evloop.ModReadWrite(conn.Fd) // 修改为监听写事件
				conn.ReadBuff.Reset()
			}
		}
	}

	return
}

func (srv *Server) HandleWrite(fd int) (err error) {

	var (
		writeN int
	)
	if conn, ok := srv.Conns[fd]; ok {
		if conn.WriteBuff.Len() > 0 {
			writeN, err = syscall.Write(fd, conn.WriteBuff.Bytes())
			if err != nil {
				if err == syscall.EAGAIN {
					err = nil
					fmt.Println("WRITE EAGIN")
					return
				}
				fmt.Println("Write err: ", err.Error())
				_ = srv.CloseFd(fd)
				return
			}

			if writeN < 0 {
				fmt.Println("Write 0")
				return
			}

			if writeN == conn.WriteBuff.Len() {
				conn.WriteBuff.Reset()
				_ = srv.evloop.ModRead(conn.Fd)
			}

			if writeN < conn.WriteBuff.Len() {
				conn.WriteBuff.Truncate(writeN)
				srv.evloop.ModReadWrite(conn.Fd)
			}
		}
	}

	return
}

func (srv *Server) CloseFd(fd int) (err error) {
	err = syscall.Close(fd)
	err = srv.evloop.Remove(fd)
	delete(srv.Conns, fd)
	return
}
