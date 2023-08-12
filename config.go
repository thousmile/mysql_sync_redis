package main

import (
	"github.com/spf13/viper"
	"log"
)

// Config 全局配置配置文件
var Config *AppConfig

func init() {
	viper.SetDefault("appName", "mysql-sync-redis")

	viper.SetDefault(
		"mysql",
		MysqlConfig{
			Addr:     "127.0.0.1:3306",
			Username: "root",
			Password: "root",
		},
	)

	viper.SetDefault(
		"redis",
		RedisConfig{
			Addrs: []string{"127.0.0.1:3306"},
			DB:    0,
		},
	)

	viper.SetDefault(
		"rules",
		map[string]SyncRule{
			"canal.canal_test": SyncRule{
				TableId:      "id",
				RedisKey:     "canal",
				RedisKeyType: "hash",
			},
		},
	)

	viper.SetConfigName("config")                // name of config file (without extension)
	viper.SetConfigType("yaml")                  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/appname/")         // 查找配置文件所在路径
	viper.AddConfigPath("$HOME/.appname")        // 多次调用AddConfigPath，可以添加多个搜索路径
	viper.AddConfigPath(".")                     // optionally look for config in the working directory
	viper.AddConfigPath("../")                   // optionally look for config in the working directory
	viper.AddConfigPath("./conf/")               // 还可以在工作目录中搜索配置文件
	if err := viper.ReadInConfig(); err != nil { // Handle errors reading the config file
		log.Panicf("Fatal error config file: %v \n", err)
	}
	if err := viper.Unmarshal(&Config); err != nil {
		log.Panicf("Fatal error config file: %v \n", err)
	}
}

type AppConfig struct {
	// 服务名称 ，默认: mysql-sync-redis
	AppName string `yaml:"appName" json:"appName"`

	// mysql 的配置
	Mysql MysqlConfig `yaml:"mysql" json:"mysql"`

	// redis 的配置
	Redis RedisConfig `yaml:"redis" json:"redis"`

	// 同步的规则
	Rules map[string]SyncRule `yaml:"rules" json:"rules"`
}

type SyncRule struct {
	// 表用作ID 的名称
	TableId string `yaml:"tableId" json:"tableId"`

	// redis 的key
	RedisKey string `yaml:"redisKey" json:"redisKey"`

	// redis 的 key 类型。string or hash
	RedisKeyType string `yaml:"redisKeyType" json:"redisKeyType"`
}

type MysqlConfig struct {
	// mysql 地址。默认: 127.0.0.1:3306
	Addr string `yaml:"addr" json:"addr"`

	// 用户名，默认: root
	Username string `yaml:"username" json:"username"`

	// 密码，默认: root
	Password string `yaml:"password" json:"password"`
}

type RedisConfig struct {
	// redis 地址。默认: 127.0.0.1:6379
	Addrs []string `yaml:"addrs" json:"addrs"`

	// 密码，默认: 空
	Password string `yaml:"password" json:"password"`

	// 库索引，默认: 0
	DB int `yaml:"db" json:"db"`

	// Sentinel 模式。
	MasterName string `yaml:"masterName" json:"masterName"`
}
