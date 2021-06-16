package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/aizsfgk/kimego/lib/log"
	"github.com/aizsfgk/kimego/lib/log/log4go"
)

// 应用层流量转发系统

var (
	help        *bool   = flag.Bool("h", false, "show help")
	confRoot    *string = flag.String("c", "./conf", "root path of config")
	logPath     *string = flag.String("l", "./log", "dir path of log")
	debugLog    *bool   = flag.Bool("d", false, "show debug lgo")
	stdOut      *bool   = flag.Bool("s", false, "show log in stdout")
	showVersion *bool   = flag.Bool("v", false, "show version of hulu")
	showVerbose *bool   = flag.Bool("V", false, "show verbose of hulu")
)

var version string
var commit string

func main() {

	var (
		err      error
		logLevel string
	)

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	// 编译的时候，将版本号编译进二进制文件
	if *showVersion {
		fmt.Printf("hulu version: %s\n", version)
		return
	}

	if *showVerbose {
		fmt.Printf("hulu version: %s\n", version)
		fmt.Printf("hulu commit: %s\n", commit)
		fmt.Printf("Go version: %s\n", runtime.Version())
		return
	}

	if *debugLog {
		logLevel = "DEBUG"

		// debug service

	} else {
		logLevel = "INFO"
	}

	// 日志初始化
	log4go.SetLogBufferLength(10000)
	log4go.SetLogFormat(log4go.FORMAT_DEFAULT)
	err = log.Init("hulu", logLevel, *logPath, *stdOut, "midnight", 7)
	if err != nil {
		fmt.Println("hulu: err in log.Init(): %s", err.Error())
		return
	}

	log.Logger.Info("hulu[version:%s] start", version)

	// 加载配置

	// 调试配置

	// 启动服务

	// 等待logger finish
	time.Sleep(1 * time.Second)
	log.Logger.Close()
}
