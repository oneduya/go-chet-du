/**
 * Created by lock
 * Date: 2019-08-12
 * Time: 15:44
 */
package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"gochat/config"
	"gochat/proto"
	"gochat/tools"
	"strings"
	"time"
)

var RedisClient *redis.Client
var RedisSessClient *redis.Client

/*启动操作Redis客户端*/
func (logic *Logic) InitPublishRedisClient() (err error) {
	//redis连接参数
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Common.CommonRedis.RedisAddress,
		Password: config.Conf.Common.CommonRedis.RedisPassword,
		Db:       config.Conf.Common.CommonRedis.Db,
	}
	//创建redis客户端实例，redis结构体放在tools包中，调用Api创建
	RedisClient = tools.GetRedisInstance(redisOpt)
	//检查连接
	if pong, err := RedisClient.Ping().Result(); err != nil {
		logrus.Infof("RedisCli Ping Result pong: %s,  err: %s", pong, err)
	}
	//this can change use another redis save session data
	//这里搞不懂为啥要换一个对象保存
	RedisSessClient = RedisClient
	return err
}

func (logic *Logic) InitRpcServer() (err error) {
	var network, addr string
	// a host multi port case
	rpcAddressList := strings.Split(config.Conf.Logic.LogicBase.RpcAddress, ",")
	for _, bind := range rpcAddressList {
		if network, addr, err = tools.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitLogicRpc ParseNetwork error : %s", err.Error())
		}
		logrus.Infof("logic start run at-->%s:%s", network, addr)
		go logic.createRpcServer(network, addr)
	}
	return
}

func (logic *Logic) createRpcServer(network string, addr string) {
	//首先创建新的服务器对象
	s := server.NewServer()
	//在server对象里面创建一个serverplugin.EtcdV3RegisterPlugin的插件
	//插件必须在注册服务之前添加到Server中，否则插件没有办法获取注册的服务的信息，可以在注册方法RegisterName中看到调用了插件的方法
	logic.addRegistryPlugin(s, network, addr)
	// serverId must be unique
	//err := s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(RpcLogic), fmt.Sprintf("%s", config.Conf.Logic.LogicBase.ServerId))
	//这里是注册方法
	err := s.RegisterName(config.Conf.Common.CommonEtcd.ServerPathLogic, new(RpcLogic), fmt.Sprintf("%s", logic.ServerId))
	if err != nil {
		logrus.Errorf("register error:%s", err.Error())
	}
	s.RegisterOnShutdown(func(s *server.Server) {
		s.UnregisterAll()
	})
	s.Serve(network, addr)
}

func (logic *Logic) addRegistryPlugin(s *server.Server, network string, addr string) {
	//定义EtcdV3RegisterPlugin插件，
	r := &serverplugin.EtcdV3RegisterPlugin{
		//本机监听地址
		ServiceAddress: network + "@" + addr,
		//etcd集群地址
		EtcdServers: []string{config.Conf.Common.CommonEtcd.Host},
		//服务前缀
		BasePath: config.Conf.Common.CommonEtcd.BasePath,
		//用来更新服务的TPS，不太了解
		Metrics: metrics.NewRegistry(),
		//服务刷新间隔，目前不太了解
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logrus.Fatal(err)
	}
	s.Plugins.Add(r)
}

func (logic *Logic) RedisPublishChannel(serverId string, toUserId int, msg []byte) (err error) {
	redisMsg := proto.RedisMsg{
		Op:       config.OpSingleSend,
		ServerId: serverId,
		UserId:   toUserId,
		Msg:      msg,
	}
	redisMsgStr, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPublishChannel Marshal err:%s", err.Error())
		return err
	}
	redisChannel := config.QueueName
	if err := RedisClient.Publish(redisChannel, redisMsgStr).Err(); err != nil {
		logrus.Errorf("logic,RedisPublishChannel err:%s", err.Error())
		return err
	}
	return
}

func (logic *Logic) RedisPublishRoomInfo(roomId int, count int, RoomUserInfo map[string]string, msg []byte) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:           config.OpRoomSend,
		RoomId:       roomId,
		Count:        count,
		Msg:          msg,
		RoomUserInfo: RoomUserInfo,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPublishRoomInfo redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.Publish(config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("logic,RedisPublishRoomInfo redisMsg error : %s", err.Error())
		return
	}
	return
}

func (logic *Logic) RedisPushRoomCount(roomId int, count int) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:     config.OpRoomCountSend,
		RoomId: roomId,
		Count:  count,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomCount redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.Publish(config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomCount redisMsg error : %s", err.Error())
		return
	}
	return
}

func (logic *Logic) RedisPushRoomInfo(roomId int, count int, roomUserInfo map[string]string) (err error) {
	var redisMsg = &proto.RedisMsg{
		Op:           config.OpRoomInfoSend,
		RoomId:       roomId,
		Count:        count,
		RoomUserInfo: roomUserInfo,
	}
	redisMsgByte, err := json.Marshal(redisMsg)
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomInfo redisMsg error : %s", err.Error())
		return
	}
	err = RedisClient.Publish(config.QueueName, redisMsgByte).Err()
	if err != nil {
		logrus.Errorf("logic,RedisPushRoomInfo redisMsg error : %s", err.Error())
		return
	}
	return
}

func (logic *Logic) getRoomUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

func (logic *Logic) getRoomOnlineCountKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisRoomOnlinePrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}

func (logic *Logic) getUserKey(authKey string) string {
	var returnKey bytes.Buffer
	returnKey.WriteString(config.RedisPrefix)
	returnKey.WriteString(authKey)
	return returnKey.String()
}
