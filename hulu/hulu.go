package main

import (
	"flag"
	"fmt"
	"path"
	"runtime"
	"time"

	"github.com/aizsfgk/kimego/hulu/hulu_config/hulu_conf"
	"github.com/aizsfgk/kimego/hulu/hulu_debug"
	"github.com/aizsfgk/kimego/hulu/hulu_server"

	"github.com/aizsfgk/kimego/lib/log"
	"github.com/aizsfgk/kimego/lib/log/log4go"
)

// 应用层流量转发系统

var (
	help        = flag.Bool("h", false, "show help")
	confRoot    = flag.String("c", "./conf", "root path of config")
	logPath     = flag.String("l", "./log", "dir path of log")
	debugLog    = flag.Bool("d", false, "show debug lgo")
	stdOut      = flag.Bool("s", false, "show log in stdout")
	showVersion = flag.Bool("v", false, "show version of hulu")
	showVerbose = flag.Bool("V", false, "show verbose of hulu")
)

var version string
var commit string

func main() {

	var (
		err      error
		logLevel string
		config   hulu_conf.HuluConfig
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
		hulu_debug.DebugIsOpen = true

	} else {
		logLevel = "INFO"
		hulu_debug.DebugIsOpen = false
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
	confPath := path.Join(*confRoot, "hulu.conf")
	config, err = hulu_conf.HuluConfigLoad(confPath, *confRoot)
	if err != nil {
		log.Logger.Error("main() in hulu_conf.HuluConfigLoad(): %s", err.Error())
		return
	}

	// 调试配置

	// 启动服务
	hulu_server.StartUp(config, version, *confRoot)

	// 等待logger finish
	time.Sleep(1 * time.Second)
	log.Logger.Close()
}
