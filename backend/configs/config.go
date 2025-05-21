package configs

import (
	"path/filepath"
	"sync"
)

type Config struct {
	App
	Log
}

type App struct {
	Mode       string `json:"mode"`
	Title      string `json:"title"`
	MacTitle   string `json:"mac_title"`
	MacMessage string `json:"mac_message"`
}

type Log struct {
	LogDir      string `json:"log_dir"`
	LogFileName string `json:"log_file_name"`
	Level       string `json:"level"`
	MaxSize     int    `json:"max_size"`
	MaxBackups  int    `json:"max_backups"`
	MaxAge      int    `json:"max_age"`
	Compress    bool   `json:"compress"`
	LocalTime   bool   `json:"local_time"`
}

var (
	cfg  *Config
	once sync.Once
)

func initConfig() *Config {
	return &Config{
		App: App{
			Mode:       "dev", // dev、prod
			Title:      "小绿书爬虫",
			MacTitle:   "wxGraphCrawler",
			MacMessage: "专用于抓取微信公众号“图片/文字”类型（俗称：小绿书）中的图片小工具 \r\n @Copyright 2025 by Alex",
		},
		Log: Log{
			LogDir:      filepath.Join(".wxGraphCrawler", "logs"),
			LogFileName: "crawl.log",
			Level:       "info", // debug、info、warn、error
			MaxSize:     1,
			MaxBackups:  5,
			MaxAge:      30,
			Compress:    false,
			LocalTime:   true,
		},
	}
}

func GetConfig() *Config {
	once.Do(func() {
		cfg = initConfig()
	})
	return cfg
}
