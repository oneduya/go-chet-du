/**
 * Created by lock
 * Date: 2019-08-09
 * Time: 18:25
 */
package logic

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gochat/config"
	"runtime"
)

type Logic struct {
	ServerId string
}

func New() *Logic {
	return new(Logic)
}

func (logic *Logic) Run() {
	//首先获取config中的内容，在导入包的时候config.go中的init方法就可以读取配置文件中的内容
	logicConfig := config.Conf.Logic
	//设置P的最大个数
	runtime.GOMAXPROCS(logicConfig.LogicBase.CpuNum)
	logic.ServerId = fmt.Sprintf("logic-%s", uuid.New().String())
	//init publish redis
	if err := logic.InitPublishRedisClient(); err != nil {
		logrus.Panicf("logic init publishRedisClient fail,err:%s", err.Error())
	}

	//init rpc server
	/*初始化rpc服务端，在服务端中要加入EtcdRegisterPlugin插件*/
	if err := logic.InitRpcServer(); err != nil {
		logrus.Panicf("logic init rpc server fail")
	}
}
