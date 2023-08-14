package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/redis/go-redis/v9"
	"github.com/siddontang/go-log/log"
	"golang.org/x/exp/maps"
	"strings"
)

var (
	rdbClient redis.UniversalClient
)

func main() {
	options := redis.UniversalOptions{
		Addrs: Config.Redis.Addrs,
		DB:    Config.Redis.DB,
	}
	if len(strings.TrimSpace(Config.Redis.Password)) > 0 {
		options.Password = Config.Redis.Password
	}
	if len(strings.TrimSpace(Config.Redis.MasterName)) > 0 {
		options.MasterName = Config.Redis.MasterName
	}
	rdbClient = redis.NewUniversalClient(&options)

	cfg := canal.NewDefaultConfig()
	cfg.Addr = Config.Mysql.Addr
	// CREATE USER canal IDENTIFIED BY 'canal';
	// GRANT RELOAD,SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'canal'@'%';
	// FLUSH PRIVILEGES;
	cfg.User = Config.Mysql.Username
	cfg.Password = Config.Mysql.Password
	cfg.Dump.ExecutionPath = ""
	c, err := canal.NewCanal(cfg)
	if err != nil {
		log.Fatal(err)
	}
	// Register a handler to handle RowsEvent
	c.SetEventHandler(&MyEventHandler{})
	// Start canal
	if err = c.Run(); err != nil {
		log.Fatal(err.Error())
	}
}

type MyEventHandler struct {
	canal.DummyEventHandler
}

func (h *MyEventHandler) OnRow(e *canal.RowsEvent) error {
	tableSchema := fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Name)
	if tv, ok1 := Config.Rules[tableSchema]; ok1 {
		objList := make(map[string]interface{}, len(e.Rows))
		isHash := tv.RedisKeyType == "hash"
		for _, row := range e.Rows {
			obj := make(map[string]interface{}, len(row))
			for i, column := range e.Table.Columns {
				obj[column.Name] = row[i]
			}
			if id, ok2 := obj[tv.TableId]; ok2 {
				idStr := fmt.Sprintf("%v", id)
				valJson, _ := json.Marshal(obj)
				objList[idStr] = valJson
			}
		}
		// 判断当前操作是 删除
		if e.Action == canal.DeleteAction {
			// 判断需要删除的 redis key 类型
			if isHash {
				hashKeys := maps.Keys(objList)
				rdbClient.HDel(context.Background(), tv.RedisKey, hashKeys...)
			} else {
				for idStr, _ := range objList {
					redisKey := fmt.Sprintf("%s:%s", tv.RedisKey, idStr)
					_, err := rdbClient.Del(context.Background(), redisKey).Result()
					if err != nil {
						log.Errorln(err)
					}
				}
			}
		} else {
			// 判断需要存储的 redis key 类型
			if isHash {
				_, err := rdbClient.HSet(context.Background(), tv.RedisKey, objList).Result()
				if err != nil {
					log.Errorln(err)
				}
			} else {
				for idStr, valJson := range objList {
					redisKey := fmt.Sprintf("%s:%s", tv.RedisKey, idStr)
					_, err := rdbClient.Set(context.Background(), redisKey, valJson, redis.KeepTTL).Result()
					if err != nil {
						log.Errorln(err)
					}
				}
			}
		}
	}
	return nil
}

func (h *MyEventHandler) String() string {
	return "MyEventHandler"
}
