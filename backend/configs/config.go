package configs

import (
	"sync"
)

type Config struct {
	App
	Log
}

type App struct {
	Mode string `json:"mode"`
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
			Mode: "dev",
		},
		Log: Log{
			LogDir:      "./.wxGraphCrawler/logs/",
			LogFileName: "crawl.log",
			Level:       "debug", // debug、info、warn、error
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
