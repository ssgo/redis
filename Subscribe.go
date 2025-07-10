package redis

import (
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

func (rd *Redis) Subscribe(name string, reset func(), received func([]byte)) bool {
	rd.subs[name] = &SubCallbacks{reset: reset, received: received}
	if rd.subConn != nil {
		//rd.subLock.Lock()
		err := rd.subConn.Subscribe(name)
		//rd.subLock.Unlock()
		if err != nil {
			rd.logger.Error(err.Error(), "subscribeName", name)
		} else {
			return true
		}
	}
	return false
}

func (rd *Redis) Unsubscribe(name string) bool {
	delete(rd.subs, name)
	if rd.subConn != nil {
		//rd.subLock.Lock()
		err := rd.subConn.Unsubscribe(name)
		//rd.subLock.Unlock()
		if err != nil {
			rd.logger.Error(err.Error(), "subscribeName", name)
		} else {
			return true
		}
	}
	return false
}

func (rd *Redis) Start() {
	//rd.subLock = sync.Mutex{}
	if rd.subs == nil {
		rd.subs = make(map[string]*SubCallbacks)
	}
	rd.SubRunning = true
	subStartChan := make(chan bool)
	go rd.receiveSub(subStartChan)
	<-subStartChan
}

func (rd *Redis) receiveSub(subStartChan chan bool) {
	for {
		if !rd.SubRunning {
			break
		}

		// 开始接收订阅数据
		if rd.subConn == nil {
			conn, err := rd.GetConnection()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			rd.subConn = &redis.PubSubConn{Conn: conn}
			// 重新订阅
			if len(rd.subs) > 0 {
				subs := make([]interface{}, 0)
				for k := range rd.subs {
					subs = append(subs, k)
				}
				//fmt.Println("     @@@@@", subs)
				err = rd.subConn.Subscribe(subs...)
				if err != nil {
					rd.subConn.Close()
					rd.subConn = nil
					time.Sleep(time.Second)
					continue
				}
				// 重新连接时调用重置数据的回掉
				for _, v := range rd.subs {
					if v.reset != nil {
						v.reset()
					}
				}
			}
		}

		if subStartChan != nil {
			subStartChan <- true
			subStartChan = nil
		}

		for {
			isErr := false
			receiveObj := rd.subConn.Receive()
			//receiveObj := rd.subConn.ReceiveWithTimeout(50*time.Millisecond)
			//fmt.Println("  >>>", receiveObj)
			switch v := receiveObj.(type) {
			case redis.Message:
				callback := rd.subs[v.Channel]
				if callback.received != nil {
					callback.received(v.Data)
				}
			case redis.Subscription:
			case redis.Pong:
			case error:
				if strings.Contains(v.Error(), "i/o timeout") {
					break
				}
				if !strings.Contains(v.Error(), "connection closed") && !strings.Contains(v.Error(), "use of closed network connection") {
					rd.logger.Error(v.Error())
				}
				if rd.subConn != nil {
					_ = rd.subConn.Close()
					rd.subConn = nil
				}
				isErr = true
				break
			}
			if isErr {
				break
			}
			if !rd.SubRunning {
				break
			}
		}
		if !rd.SubRunning {
			break
		}
	}
	if rd.subStopChan != nil {
		rd.subStopChan <- true
	}
}

func (rd *Redis) Stop() {
	if rd.SubRunning {
		rd.subStopChan = make(chan bool)
		rd.SubRunning = false
		if rd.subConn != nil {
			// 取消订阅
			if len(rd.subs) > 0 {
				subs := make([]interface{}, 0)
				for k := range rd.subs {
					subs = append(subs, k)
				}
				_ = rd.subConn.Unsubscribe()
			}
			// 读一次再关闭可以防止Close时阻塞
			_ = rd.subConn.ReceiveWithTimeout(50 * time.Millisecond)
			_ = rd.subConn.Close()
			rd.subConn = nil
		}
		<-rd.subStopChan
		rd.subStopChan = nil
	}
}
