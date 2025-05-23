package configs

import (
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	App
	Log
	DB
}

type App struct {
	Mode         string `json:"mode"`
	Title        string `json:"title"`
	MacTitle     string `json:"mac_title"`
	MacMessage   string `json:"mac_message"`
	LinuxMessage string `json:"linux_message"`
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

type DB struct {
	SQLite3FilePath    string `json:"sqlite3_file_path"`    // SQLite3数据库文件路径
	MaxOpenConnections int    `json:"max_open_connections"` // 最大打开连接数
	MaxIdleConnections int    `json:"max_idle_connections"` // 最大空闲连接数
	MaxLifeSeconds     int    `json:"max_life_seconds"`     // 连接的最大生命周期，单位秒，设置为0表示不限制
}

var (
	cfg  *Config
	once sync.Once
)

func initConfig() *Config {
	appDescription := "wxGraphCrawler 是一个专门用于抓取微信公众号“图片/文字”类型（俗称：小绿书）图片的小工具 \r\n" +
		"Homepage：https://github.com/pudongping/wx-graph-crawl \r\n" +
		"@Copyright " + time.Now().Format("2006") + " by Alex"

	return &Config{
		App: App{
			Mode:         "dev", // dev、prod
			Title:        "小绿书爬虫",
			MacTitle:     "wxGraphCrawler",
			MacMessage:   appDescription,
			LinuxMessage: appDescription,
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
		DB: DB{
			SQLite3FilePath:    filepath.Join(".wxGraphCrawler", "data", "wx_graph_crawler.db"),
			MaxOpenConnections: 25,
			MaxIdleConnections: 100,
			MaxLifeSeconds:     300, // 单位，秒；设置为0表示不限制
		},
	}
}

func GetConfig() *Config {
	once.Do(func() {
		cfg = initConfig()
	})
	return cfg
}
