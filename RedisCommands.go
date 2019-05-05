package redis

func StringsToInterfaces(in []string) []interface{} {
	a := make([]interface{}, len(in))
	for i, v := range in {
		a[i] = v
	}
	return a
}

func (rd *Redis) DEL(keys ...string) int {
	return rd.Do("DEL", StringsToInterfaces(keys)...).Int()
}
func (rd *Redis) EXISTS(key string) bool {
	return rd.Do("EXISTS " + key).Bool()
}
func (rd *Redis) EXPIRE(key string, second int) bool {
	if second > 315360000 {
		return rd.Do("EXPIREAT "+key, second).Bool()
	} else {
		return rd.Do("EXPIRE "+key, second).Bool()
	}
}
func (rd *Redis) KEYS(patten string) []string {
	return rd.Do("KEYS " + patten).Strings()
}

func (rd *Redis) GET(key string) *Result {
	return rd.Do("GET " + key)
}
func (rd *Redis) SET(key string, value interface{}) bool {
	return rd.Do("SET "+key, value).Bool()
}
func (rd *Redis) SETEX(key string, seconds int, value interface{}) bool {
	return rd.Do("SETEX "+key, seconds, value).Bool()
}
func (rd *Redis) SETNX(key string, value interface{}) bool {
	return rd.Do("SETNX "+key, value).Bool()
}
func (rd *Redis) GETSET(key string, value interface{}) *Result {
	return rd.Do("GETSET "+key, value)
}

func (rd *Redis) INCR(key string) int64 {
	return rd.Do("INCR " + key).Int64()
}
func (rd *Redis) DECR(key string) int64 {
	return rd.Do("DECR " + key).Int64()
}

func (rd *Redis) MGET(keys ...string) []Result {
	return rd.Do("MGET", StringsToInterfaces(keys)...).Results()
}
func (rd *Redis) MSET(keyAndValues ...interface{}) bool {
	return rd.Do("MSET", keyAndValues...).Bool()
}

func (rd *Redis) HGET(key, field string) *Result {
	return rd.Do("HGET "+key, field)
}
func (rd *Redis) HSET(key, field string, value interface{}) bool {
	return rd.Do("HSET "+key, field, value).Bool()
}
func (rd *Redis) HSETNX(key, field string, value interface{}) bool {
	return rd.Do("HSETNX "+key, field, value).Bool()
}
func (rd *Redis) HMGET(key string, fields ...string) []Result {
	return rd.Do("HMGET", append(append([]interface{}{}, key), StringsToInterfaces(fields)...)...).Results()
}
func (rd *Redis) HGETALL(key string) map[string]*Result {
	return rd.Do("HGETALL " + key).ResultMap()
}
func (rd *Redis) HMSET(key string, fieldAndValues ...interface{}) bool {
	return rd.Do("HMSET", append(append([]interface{}{}, key), fieldAndValues...)...).Bool()
}
func (rd *Redis) HKEYS(key string) []string {
	return rd.Do("HKEYS " + key).Strings()
}
func (rd *Redis) HLEN(key string) int {
	return rd.Do("HLEN " + key).Int()
}
func (rd *Redis) HDEL(key string, fields ...string) int {
	return rd.Do("HDEL", append(append([]interface{}{}, key), StringsToInterfaces(fields)...)...).Int()
}
func (rd *Redis) HEXISTS(key, field string) bool {
	return rd.Do("HEXISTS "+key, field).Bool()
}
func (rd *Redis) HINCR(key, field string) int64 {
	return rd.Do("HINCRBY "+key, field, 1).Int64()
}
func (rd *Redis) HDECR(key, field string) int64 {
	return rd.Do("HDECRBY "+key, field, 1).Int64()
}

func (rd *Redis) LPUSH(key string, values ...string) int {
	return rd.Do("LPUSH", append(append([]interface{}{}, key), StringsToInterfaces(values)...)...).Int()
}
func (rd *Redis) RPUSH(key string, values ...string) int {
	return rd.Do("RPUSH", append(append([]interface{}{}, key), StringsToInterfaces(values)...)...).Int()
}
func (rd *Redis) LPOP(key string) *Result {
	return rd.Do("LPOP " + key)
}
func (rd *Redis) RPOP(key string) *Result {
	return rd.Do("RPOP " + key)
}
func (rd *Redis) LLEN(key string) int {
	return rd.Do("LLEN " + key).Int()
}
func (rd *Redis) LRANGE(key string, start, stop int) []Result {
	return rd.Do("LRANGE "+key, start, stop).Results()
}
