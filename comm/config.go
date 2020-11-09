package uller

import (
	"github.com/go-ini/ini"
)

type Config struct {
	Path					string				// 程序运行目录
	IP						string				// 本机IP
	IniReader				*ini.File			// ini文件读取器
}

func ConfigInstance(configConfig ...Config) *Config{
	config := Config{}
	if len(configConfig) == 1{
		config.Path = configConfig[0].Path
		config.IP = configConfig[0].Path
		config.IniReader = configConfig[0].IniReader
	}else{
		config.Path = GetCurrentDirectory()
		config.IP = GetLocalIP()
		cfg, err := ini.Load(config.Path + "/config.ini")
		if err != nil{
			panic("ini file not exist")
		}
		config.IniReader = cfg
	}
	return &config
}