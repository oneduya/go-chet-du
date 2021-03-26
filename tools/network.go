/**
 * Created by lock
 * Date: 2019-08-12
 * Time: 16:00
 */
package tools

import (
	"fmt"
	"strings"
)

const (
	networkSplit = "@"
)

/*处理网络地址的函数，将配置中的地址按@符分成连接协议和地址
如”tcp@127.0.0.1:6900“，分成network=”tcp“和”127.0.0.1:6900“
*/
func ParseNetwork(str string) (network, addr string, err error) {
	//判断地址中有没有@分隔符
	if idx := strings.Index(str, networkSplit); idx == -1 {
		//没有就报错并且给出正确的格式
		err = fmt.Errorf("addr: \"%s\" error, must be network@tcp:port or network@unixsocket", str)
		return
	} else {
		//按照分隔符分开连接协议和地址
		network = str[:idx]
		addr = str[idx+1:]
		return
	}
}
