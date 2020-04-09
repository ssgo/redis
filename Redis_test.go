package redis_test

import (
	"github.com/ssgo/config"
	"github.com/ssgo/redis"
	"os"
	"testing"
	"time"
)

type userInfo struct {
	Id    int
	Name  string
	Phone string
	Time  time.Time
}

func TestBase(t *testing.T) {
	os.Setenv("redis_test2", "redis://:@localhost:6379/2?timeout=10ms&logSlow=10us&database=4")
	config.ResetConfigEnv()

	redis := redis.GetRedis("test", nil)
	redis.Config.LogSlow = -1
	if redis.Error != nil {
		t.Error("GetRedis error", redis)
		return
	}
	redis.DEL("redisName", "redisUser", "redisIds")

	r := redis.GET("redisNotExists")
	if r.Error != nil && r.String() != "" || r.Int() != 0 {
		t.Error("GET NotExists", r, r.String(), r.Int())
	}

	exists := redis.EXISTS("redisName")
	if exists {
		t.Error("EXISTS", exists)
	}

	redis.SET("redisName", "12345")

	r = redis.GETSET("redisName", 12345)
	if r.Error != nil && r.String() != "12345" {
		t.Error("String", r)
	}
	if r.Int() != 12345 {
		t.Error("Int", r)
	}
	if r.Float() != 12345 {
		t.Error("Float", r)
	}
	if r.Bool() != false {
		t.Error("Bool", r)
	}

	exists = redis.EXISTS("redisName")
	if !exists {
		t.Error("EXISTS", exists)
	}

	r = redis.GET("redisName")
	if r.Error != nil && r.String() != "12345" {
		t.Error("String", r)
	}
	if r.Int() != 12345 {
		t.Error("Int", r)
	}
	if r.Float() != 12345 {
		t.Error("Float", r)
	}
	//过期测试 测试小于315360000(3650天)的unix时间戳
	expireFlag := redis.EXPIRE("redisName", 1)
	if !expireFlag {
		t.Error("Expire Bool", expireFlag)
	}
	time.Sleep(time.Duration(2) * time.Second)
	r = redis.GET("redisName")
	if r.Int() > 0 {
		t.Error("Expired err", r)
	}

	//过期测试 测试小于3650天的unix时间戳
	redis.SET("redisName", 12)
	//2019-01-01
	expireFlag = redis.EXPIRE("redisName", 1546272000)
	if !expireFlag {
		t.Error("ExpireAt Bool", expireFlag)
	}
	r = redis.GET("redisName")
	if r.Int() > 0 {
		t.Error("ExpireAt err", r)
	}

	redis.SET("redisName", 16)
	//2030-01-01
	expireFlag = redis.EXPIRE("redisName", 1893427200)
	if !expireFlag {
		t.Error("ExpireAt Bool", expireFlag)
	}
	r = redis.GET("redisName")
	if r.Int() != 16 {
		t.Error("ExpireAt err", r)
	}

	redis.SET("redisName", 12345.67)
	r = redis.GET("redisName")
	if r.Error != nil && r.String() != "12345.67" {
		t.Error("String", r)
	}
	if r.Float() != 12345.67 {
		t.Error("Float", r)
	}
	if r.Uint64() != 12345 {
		t.Error("Uint64", r)
	}

	info := userInfo{
		Name: "aaa",
		Id:   123,
		Time: time.Now(),
	}
	redis.SET("redisUser", info)
	r = redis.GET("redisUser")
	ru := new(userInfo)
	r.To(ru)
	if r.Error != nil && ru.Name != "aaa" || ru.Id != 123 || !ru.Time.Equal(info.Time) {
		t.Error("userInfo Struct", ru)
	}

	rm := map[string]interface{}{}
	r.To(&rm)
	if rm["name"] != "aaa" || rm["id"].(float64) != 123 {
		t.Error("userInfo Map", rm)
	}

	keys := redis.KEYS("redis*")
	if len(keys) != 2 {
		t.Error("keys", keys)
	}

	redis.MSET("redisName", "Sam Lee", "redisUser", map[string]interface{}{
		"name": "BBB",
	})
	results := redis.MGET("redisName", "redisUser")
	if len(results) != 2 || results[0].String() != "Sam Lee" {
		t.Error("MGET Results", results)
	}
	ru2 := new(userInfo)
	results[1].To(ru2)
	if ru2.Name != "BBB" {
		t.Error("MGET Struct", results)
	}

	r = redis.Do("MGET", "redisName", "redisUser")
	r1 := make([]string, 0)
	r.To(&r1)
	if len(r1) != 2 || r1[0] != "Sam Lee" {
		t.Error("MGET2 Strings", r1)
	}
	//r2 := struct {
	//	RtestName string
	//	RtestUser userInfo
	//}{}
	//r.To(&r2)
	//if r2.RtestName != "Sam Lee" || r2.RtestUser.Name != "BBB" {
	//	t.Error("MGET2 Struct and Struct", u.JsonP(r2), u.JsonP(r.Strings()))
	//}
	rm2 := r.ResultMap()
	if rm2["redisName"].String() != "Sam Lee" || rm2["redisUser"].ResultMap()["name"].String() != "BBB" {
		t.Error("MGET2 ResultMap", rm2)
	}
	ra2 := r.Results()
	if ra2[0].String() != "Sam Lee" || ra2[1].ResultMap()["name"].String() != "BBB" {
		t.Error("MGET2 ResultMap", ra2)
	}

	redis.SET("redisIds", []interface{}{1, 2, 3})
	r = redis.GET("redisIds")
	ria := r.Ints()
	if ria[0] != 1 || ria[1] != 2 || ria[2] != 3 {
		t.Error("userIds Ints", ria)
	}

	num := redis.DEL("redisName", "redisUser", "redisIds")
	if num != 3 {
		t.Error("DEL", num)
	}
}

func TestConfigByName(t *testing.T) {
	redis := redis.GetRedis("localhost:6379:2", nil)
	//fmt.Println(redis.Config)
	if redis.Error != nil {
		t.Error("GetRedis error", redis)
		return
	}

	redis.SET("redisName", "12345")
	r := redis.GET("redisName")
	if r.Error != nil && r.String() != "12345" {
		t.Error("String", r)
	}

	num := redis.DEL("redisName")
	if num != 1 {
		t.Error("DEL", num)
	}
}

func TestConfigByUrl(t *testing.T) {
	redis := redis.GetRedis("redis://:@localhost:6379/2?timeout=100ms&database=3", nil)
	//fmt.Println(redis.Config)
	if redis.Error != nil {
		t.Error("GetRedis error", redis)
		return
	}

	redis.SET("redisName", "12345")
	r := redis.GET("redisName")
	if r.Error != nil && r.String() != "12345" {
		t.Error("String", r)
	}

	num := redis.DEL("redisName")
	if num != 1 {
		t.Error("DEL", num)
	}
}

func TestConfigByUrl2(t *testing.T) {
	redis := redis.GetRedis("test2", nil)
	//fmt.Println(redis.Config)
	if redis.Error != nil {
		t.Error("GetRedis error", redis)
		return
	}

	redis.SET("redisName", "12345")
	r := redis.GET("redisName")
	if r.Error != nil && r.String() != "12345" {
		t.Error("String", r)
	}

	num := redis.DEL("redisName")
	if num != 1 {
		t.Error("DEL", num)
	}
}

func TestHash(t *testing.T) {
	redis := redis.GetRedis("test", nil)
	if redis.Error != nil {
		t.Error("GetRedis error", redis)
		return
	}

	r := redis.HGET("htest", "NotExists")
	if r.String() != "" || r.Int() != 0 {
		t.Error("GET NotExists", r, r.String(), r.Int())
	}

	exists := redis.HEXISTS("htest", "Name")
	if exists {
		t.Error("HEXISTS", exists)
	}

	redis.HSET("htest", "Name", "12345")
	r = redis.HGET("htest", "Name")
	if r.String() != "12345" {
		t.Error("String", r)
	}
	if r.Int() != 12345 {
		t.Error("Int", r)
	}
	if r.Float() != 12345 {
		t.Error("Float", r)
	}

	exists = redis.HEXISTS("htest", "Name")
	if !exists {
		t.Error("HEXISTS", exists)
	}

	redis.HSET("htest", "Name", 12345.67)
	r = redis.HGET("htest", "Name")
	if r.String() != "12345.67" {
		t.Error("String", r)
	}
	if r.Float() != 12345.67 {
		t.Error("Float", r)
	}
	if r.Uint64() != 12345 {
		t.Error("Uint64", r)
	}

	u := userInfo{
		Name: "aaa",
		Id:   123,
		Time: time.Now(),
	}
	redis.HSET("htest", "User", u)
	ru := new(userInfo)
	redis.HGET("htest", "User").To(ru)
	redis.HGET("htest", "User").To(ru)
	if ru.Name != "aaa" || ru.Id != 123 || !ru.Time.Equal(u.Time) {
		t.Error("Ints", ru)
	}

	rm := map[string]interface{}{}
	redis.HGET("htest", "User").To(&rm)
	if rm["name"] != "aaa" || rm["id"].(float64) != 123 {
		t.Error("user", rm)
	}

	redis.HMSET("htest", "Name", "Sam Lee", "User", map[string]interface{}{
		"name": "BBB",
	})
	results := redis.HMGET("htest", "Name", "User")
	if len(results) != 2 || results[0].String() != "Sam Lee" {
		t.Error("HMGET", results[0])
	}
	ru2 := new(userInfo)
	results[1].To(ru2)
	if ru2.Name != "BBB" {
		t.Error("HMGET", results[1])
	}

	r = redis.Do("HMGET", "htest", "Name", "User")
	r1 := make([]string, 0)
	r.To(&r1)
	if r.Error != nil && len(r1) != 2 || r1[0] != "Sam Lee" {
		t.Error("HMGET r1", r1)
	}

	r2 := struct {
		Name string
		User userInfo
	}{}
	r.To(&r2)
	if r2.Name != "Sam Lee" && r2.User.Name != "BBB" {
		t.Error("HMGET r2", r2)
	}

	rm3 := redis.HGETALL("htest")
	if rm3["Name"].String() != "Sam Lee" || rm3["User"].ResultMap()["name"].String() != "BBB" {
		t.Error("HGETALL ResultMap", rm3)
	}

	redis.HSET("htest", "Ids", []interface{}{1, 2, 3})
	r = redis.HGET("htest", "Ids")
	ria := r.Ints()
	if ria[0] != 1 || ria[1] != 2 || ria[2] != 3 {
		t.Error("userIds Ints", ria)
	}

	keys := redis.HKEYS("htest")
	if len(keys) != 3 {
		t.Error("HKEYS", keys)
	}

	len := redis.HLEN("htest")
	if len != 3 {
		t.Error("HLEN", keys)
	}

	num := redis.DEL("htest")
	if num != 1 {
		t.Error("DEL", num)
	}
}
