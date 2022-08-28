package cache

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
	"time"
	"user/pkg/util"
)

var Redis redis.Conn

func InitRedis() {
	address := viper.GetString("redis.address")
	// 连接redis
	c, err := redis.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		util.LogrusObj.Error(err)
	}
	Redis = c
}
func c()  {
	// 连接redis
	c, err := redis.Dial("tcp", "192.168.151.158:12004")
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	} else {
		fmt.Println("Connect to redis ok")
	}
	defer c.Close()

	// 密码鉴权
	_, err = c.Do("AUTH", "cqrm123151qaz2WSX")
	if err != nil {
		fmt.Println("auth failed:", err)
	} else {
		fmt.Println("auth ok:")
	}

	// 写入数据
	_, err = c.Do("SET", "gokey", "gokeyvalue")
	if err != nil {
		fmt.Println("redis set failed:", err)
	} else {
		fmt.Println("redis set ok")
	}

	// 读取数据
	value, err := redis.String(c.Do("GET", "gokey"))
	if err != nil {
		fmt.Println("redis get failed:", err)
	} else {
		fmt.Printf("Get gokey: %v \n", value)
	}

	// 删除key
	_, err = c.Do("DEL", "gokey")
	if err != nil {
		fmt.Println("redis delelte failed:", err)
	}

	// 读取数据
	value, err = redis.String(c.Do("GET", "gokey"))
	if err != nil {
		fmt.Println("redis get failed:", err)
	} else {
		fmt.Printf("Get gokey: %v \n", value)
	}

	// 组装JSON字符串
	key := "profile"
	imap := map[string]string{"username": "666", "phonenumber": "888"}
	jsonvalue, _ := json.Marshal(imap)

	// 写入JSON字符串
	n, err := c.Do("SETNX", key, jsonvalue)
	if err != nil {
		fmt.Println(err)
	}
	if n == int64(1) {
		fmt.Println("success")
	}

	// 读取JSON字符串
	var imapGet map[string]string
	valueGet, err := redis.Bytes(c.Do("GET", key))
	if err != nil {
		fmt.Println(err)
	}

	// 解析JSON
	errShal := json.Unmarshal(valueGet, &imapGet)
	if errShal != nil {
		fmt.Println(err)
	}
	fmt.Println(imapGet["username"])
	fmt.Println(imapGet["phonenumber"])

	// 设置过期时间为6秒
	ret, _ := c.Do("EXPIRE", key, 6)
	if ret == int64(1) {
		fmt.Println("success")
	}

	// 休眠8秒
	time.Sleep(8 * time.Second)

	// 判断key是否存在
	is_key_exit, err := redis.Bool(c.Do("EXISTS", key))
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("exists or not: %v \n", is_key_exit)
	}

}