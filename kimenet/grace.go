package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"time"
)

func handleSignal() {
	sigChan = make(chan os.Signal)

	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	for {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP:
			fmt.Println("get SIGHUP")
			ParentWriteFds()
		case syscall.SIGUSR1:
			fmt.Println("get SIGUSR1")
			stop()
		default:
			fmt.Println("未知的信号")
		}
	}
}

func ParentWriteFds() {
	fmt.Println("parent-server: ", server)
	server.State = Gracing

	fmt.Println("start grace...")
	os.Remove(unixSocketFile)
	fmt.Println("start grace rm file...")


	unixAddr, err := net.ResolveUnixAddr("unix", unixSocketFile)

	fmt.Println("ResolveUnixAddr")
	if err != nil {
		panic("ResolveUnixAddr: " + err.Error())
	}

	unixLn, err := net.ListenUnix("unix", unixAddr)
	fmt.Println("net.ListenUnix")
	if err != nil {
		panic("ListenUnix: "+err.Error())
	}

	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: append([]uintptr{
			os.Stdin.Fd(),
			os.Stdout.Fd(),
			os.Stderr.Fd(),
		}),
	}

	fmt.Println("start fork")
	var args []string
	for _, v := range os.Args {
		if v != "graceKey" {
			args = append(args, v)
		}
	}

	pid, err := syscall.ForkExec(os.Args[0], append(args, "graceKey"), execSpec)
	if err != nil {
		fmt.Println("forkExec err: ", err.Error())
		return
	}
	fmt.Println("end fork")

	// Write Conn
	unixConn, err := unixLn.AcceptUnix() // 阻塞再这里了
	if err != nil {
		panic("acceptUnix: " + err.Error())
	}
	fmt.Println("unixSocket create success!!!")


	var buf []byte
	fmt.Println("server.len", len(server.Conns))
	for _, conn := range server.Conns {
		buf, err = Encode(conn) /// 对数据进行了编码

		if err != nil {
			fmt.Println("Encode err: ", err.Error())
			continue
		}
		if len(buf) == 0 {
			fmt.Println("len(buf) == 0 ")
			continue
		}
		rights := syscall.UnixRights(conn.Fd)

		// buf 表示 payload
		// rights 表示带外数据
		n, oobn, err := unixConn.WriteMsgUnix(buf, rights, nil)
		if err != nil {
			fmt.Println("oob err: ", err.Error())
			break
		}

		fmt.Println("n: ", n, "; oobn: ", oobn)
	}

	fmt.Println("PID: ", pid)
}

/*
    平滑重启的两种方式：
    <1>
      1. acceptFd = accept(), 返回的fd设置 close-on-exec 标识
      2. 使用unixSocket建立通道，将文件描述符，传递给子进程；这时候fd改变，但文件控制权不变; 事件重做
    <2>
      1. acceptFd = accept(), 返回的fd 不设置 close-on-exec 标识
      2. 这时候继承了父进程FD; 如何把Context和Session也继承过来???


 */

func ChildReceiveFds() {

	fmt.Println("child-server:", server)
	fmt.Println("in grace....")
	// read conn
	// decode
	unixAddr, err := net.ResolveUnixAddr("unix", unixSocketFile)
	if err != nil {
		fmt.Println("net.ResolveUnixAddr err ")
		return
	}

	unixConn, err :=  net.DialUnix("unix", nil, unixAddr)
	if err != nil {
		fmt.Println(" net.DialUnix err ")
		return
	}

	unixConn.SetReadDeadline(time.Now().Add(10 * time.Second))

	defer func() {
		// 使用完后，关掉这个
		_ = unixConn.Close()
	}()

	// b是payload数据; conn Session/Context
	b := make([]byte, 65336)


	// oob是带外数据; fd<这里的fd>可以批量
	oob := make([]byte, 1024)


	for {
		n, oobn, _, _, err := unixConn.ReadMsgUnix(b, oob)
		if err != nil {
			fmt.Println("unixConn.ReadMsgUnix err ,", err.Error())
			break
		}

		sCtrMsg, err := syscall.ParseSocketControlMessage(oob[:oobn]);
		if err != nil {
			fmt.Println("ParseSocketControlMessage err ,", err.Error())
			break
		}

		fds, _ := syscall.ParseUnixRights(&sCtrMsg[0])

		if len(fds) == 1 && fds[0] > 0 {
			fmt.Println("fds: ", fds)
			//syscall.Close(fds[0]) /// ??? /// 这里 FDS => 9u
			//
			// 这里传过来的是一个新的文件描述符
			//
		}

		conn, err := Decode(b[:n])
		if err != nil {
			fmt.Println("DecodeErr: ", err.Error())
			continue
		}
		if conn == nil {
			fmt.Println("continue")
			continue
		}
		conn.Fd = fds[0]
		server.Conns[conn.Fd] = conn

		_ = server.evloop.AddRead(conn.Fd)
	}
}

func stop() {
	server.State = Stop
}
