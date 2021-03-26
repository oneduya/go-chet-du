/**
 * Created by lock
 * Date: 2019-08-12
 * Time: 14:18
 */
package tools

import (
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

var RedisClientMap = map[string]*redis.Client{}
var syncLock sync.Mutex

type RedisOption struct {
	Address  string
	Password string
	Db       int
}

/*从RedisClientMap中获取redis客户端实例，如果没有就新建一个放入map中，再取出*/
func GetRedisInstance(redisOpt RedisOption) *redis.Client {
	//配置连接参数
	address := redisOpt.Address
	db := redisOpt.Db
	password := redisOpt.Password
	addr := fmt.Sprintf("%s", address)

	//连接时加锁，这里是因为go语言中的哈希map是非线程安全的，所以需要手动加锁
	syncLock.Lock()
	//如果已经有了对应地址连接的客户端，直接返回
	if redisCli, ok := RedisClientMap[addr]; ok {
		return redisCli
	}
	//新建一个redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:       addr,
		Password:   password,
		DB:         db,
		MaxConnAge: 20 * time.Second,
	})
	RedisClientMap[addr] = client
	syncLock.Unlock()
	//返回的是从map中取的
	return RedisClientMap[addr]
}
