package hulu_server

import (
	"github.com/aizsfgk/kimego/hulu/hulu_config/hulu_conf"
	"github.com/aizsfgk/kimego/lib/log"
)

func StartUp(cfg hulu_conf.HuluConfig, version string, confRoot string) error {
	log.Logger.Info("服务器启动")

	return nil
}
