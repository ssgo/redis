package redis_test

import (
	"github.com/ssgo/redis"
	"testing"
	"time"
)

func TestSub(t *testing.T) {
	aaa := ""
	bbb := ""
	rd := redis.GetRedis("test", nil)
	rd.Start()
	rd.Subscribe("aaa", nil, func(s []byte) {
		//fmt.Println("aaa", s)
		aaa = string(s)
	})
	//time.Sleep(100*time.Millisecond)
	rd.Stop()
	time.Sleep(100*time.Millisecond)
	rd.Start()
	rd.Subscribe("bbb", nil, func(s []byte) {
		//fmt.Println("bbb", s)
		bbb = string(s)
	})
	rd.PUBLISH("aaa", "111")
	rd.PUBLISH("bbb", "222")

	time.Sleep(100 * time.Millisecond)
	if aaa != "111" || bbb != "222" {
		t.Fatal("TestSub", aaa, bbb)
	}

	rd.Unsubscribe("aaa")
	rd.PUBLISH("aaa", "1111")
	rd.PUBLISH("bbb", "2222")
	time.Sleep(100 * time.Millisecond)

	if aaa != "111" || bbb != "2222" {
		t.Fatal("TestSub2", aaa, bbb)
	}
	//for i:=0; i<100; i++ {
	//	rd.PUBLISH("bbb", "2222")
	//}

	rd.Stop()

}
