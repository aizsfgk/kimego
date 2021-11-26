package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
)

var (
	server *Server
	sigChan chan os.Signal
	unixSocketFile = "/tmp/sna.sock"
)

func main()  {
	var err error
	server, err = NewServer("0.0.0.0", 9192)
	if err != nil {
		panic("NewServer:" + err.Error())
	}

	server.evloop, err = NewEventLoop()
	server.evloop.serv = server
	if err != nil {
		panic("NewWventLoop err: " + err.Error())
	}

	err = server.evloop.AddRead(server.ListenFd)
	if err != nil {
		panic("add ListenFd err: "+ err.Error())
	}

	go handleSignal()

	if isGrace() {
		ChildReceiveFds()
		fmt.Println("给父进程发送SIGUSR1信号")
		syscall.Kill(os.Getppid(), syscall.SIGUSR1)
	}


	fmt.Println(os.Getpid(), "开始处理主服务器任务")


	err = server.evloop.Poll(server.EventLoopCallback)
	if err != nil {
		fmt.Println("Poll err: ", err.Error())
	}



	err = syscall.Close(server.ListenFd) // 必须关闭这个FD; 要不底层还能监听
	if err != nil {
		fmt.Println("syscall.Close err: ", err.Error())
	}

	err = server.evloop.Close()
	if err != nil {
		fmt.Println("server.evloop.Close() err: ", err.Error())
	}
	fmt.Println("服务器停止")
	return
}

func sockAddrToString(sa syscall.Sockaddr) string {
	switch sa := (sa).(type) {
	case *syscall.SockaddrInet4:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	case *syscall.SockaddrInet6:
		return net.JoinHostPort(net.IP(sa.Addr[:]).String(), strconv.Itoa(sa.Port))
	default:
		return fmt.Sprintf("(unknow - %T)", sa)
	}
}

func isGrace() bool {
	rt := false
	for _, v := range os.Args {
		if v == "graceKey" {
			rt = true
		}
	}
	return rt
}





//func handleAcceptFd(fd int, sa syscall.Sockaddr) {
//
//	if server.State == Gracing {
//		fmt.Println("服务器在平滑重启中， 不能处理accept事务")
//		return
//	}
//
//	addr := sockAddrToString(sa)
//	fmt.Printf("accept fd : %d, addr: %s\n", fd, addr)
//
//	conn, err := NewConnection(fd, fmt.Sprintf("addr : %s", addr))
//	if err != nil {
//		fmt.Println("handleAcceptFd err: ", err.Error())
//		return
//	}
//
//	server.Conns[fd] = conn
//
//
//	for {
//		readN, err := syscall.Read(conn.Fd, conn.ReadBuff.Bytes())
//		if err != nil {
//			fmt.Println("syscall.Read err: " + err.Error())
//			delete(server.Conns, conn.Fd)
//			return
//		}
//
//		if readN == 0 {
//			syscall.Close(conn.Fd)
//			conn.State = CLOSED
//			delete(server.Conns, conn.Fd)
//			fmt.Println("Conn IsClosed")
//			return
//		}
//
//		if readN > 0 {
//			fmt.Printf("get msg %d, content: %s\n", readN, string(conn.ReadBuff.Bytes()[:readN]))
//
//			syscall.Write(conn.Fd, conn.ReadBuff.Bytes()[:readN])
//		}
//	}
//
//}
//
//
//func HandleRevFd(conn *Connection) {
//	fmt.Println("开始-》 HandleRevFd")
//	for {
//		readN, err := syscall.Read(conn.Fd, conn.ReadBuff.Bytes())
//		if err != nil {
//			fmt.Println("HandleRevFd syscall.Read err: " + err.Error())
//			return
//		}
//
//		if readN == 0 {
//			syscall.Close(conn.Fd)
//			conn.State = CLOSED
//			delete(server.Conns, conn.Fd)
//			fmt.Println("HandleRevFd Conn IsClosed")
//			return
//		}
//
//		if readN > 0 {
//			fmt.Printf("get msg %d, content: %s\n", readN, string(conn.ReadBuff.Bytes()[:readN]))
//			delete(server.Conns, conn.Fd)
//
//			syscall.Write(conn.Fd, conn.ReadBuff.Bytes()[:readN])
//		}
//	}
//}


