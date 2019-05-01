package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ssgo/log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/ssgo/config"
	"github.com/ssgo/u"
)

var passwords = map[string]string{}

type Config struct {
	Host         string
	Password     string
	DB           int
	MaxActive    int
	MaxIdles     int
	IdleTimeout  int
	ConnTimeout  int
	ReadTimeout  int
	WriteTimeout int
	LogSlow      int
}

func (conf *Config) Dsn() string {
	return fmt.Sprintf("%s:****:%d", conf.Host, conf.DB)
}

type Redis struct {
	pool        *redis.Pool
	ReadTimeout int
	Config      *Config
	logger      *log.Logger
	Error       error
}

// var settedKey = []byte("vpL54DlR2KG{JSAaAX7Tu;*#&DnG`M0o")
// var settedIv = []byte("@z]zv@10-K.5Al0Dm`@foq9k\"VRfJ^~j")
var settedKey = []byte("?GQ$0K0GgLdO=f+~L68PLm$uhKr4'=tV")
var settedIv = []byte("VFs7@sK61cj^f?HZ")
var keysSetted = false

func SetEncryptKeys(key, iv []byte) {
	if !keysSetted {
		settedKey = key
		settedIv = iv
		keysSetted = true
	}
}

//var enabledLogs = true
//
//func EnableLogs(enabled bool) {
//	enabledLogs = enabled
//}

var redisConfigs = make(map[string]*Config)
var redisInstances = make(map[string]*Redis)

func GetRedis(name string, logger *log.Logger) *Redis {
	if logger == nil {
		logger = log.DefaultLogger
	}

	if redisInstances[name] != nil {
		return copyByLogger(redisInstances[name], logger)
	}

	if len(redisConfigs) == 0 {
		errs := config.LoadConfig("redis", &redisConfigs)
		if errs != nil {
			for _, err := range errs {
				logger.Error(err.Error())
			}
		}
	}

	fullName := name
	// config name support Host:Port
	args := strings.Split(name, ":")
	db := 0
	if len(args) > 1 {
		arg1, err := strconv.Atoi(args[1])
		if err == nil && arg1 > 0 && arg1 <= 15 {
			name = args[0]
			db = arg1
		}
	}

	conf := redisConfigs[name]
	if conf == nil {
		conf = new(Config)
		redisConfigs[name] = conf

		if len(args) > 1 {
			arg1, err := strconv.Atoi(args[1])
			if err == nil && arg1 > 0 && arg1 <= 15 {
				conf.DB = arg1
			} else {
				conf.Host = args[0] + ":" + args[1]
			}
		}
	}

	for i := 2; i < len(args); i++ {
		arg2, err := strconv.Atoi(args[i])
		if err == nil {
			if arg2 > 0 && arg2 <= 15 {
				conf.DB = arg2
			} else if arg2 > 15 && arg2 < 86400000 {
				if conf.ConnTimeout == 0 {
					conf.ConnTimeout = arg2
				} else {
					conf.ReadTimeout = arg2
					conf.WriteTimeout = arg2
				}
			} else if arg2 == -1 {
				conf.ReadTimeout = -1
			} else {
				conf.Password = args[i]
			}
		} else {
			conf.Password = args[i]
		}
	}

	passId := u.UniqueId()
	passwords[passId] = conf.Password
	conf.Password = passId

	if conf.Host == "" {
		conf.Host = "127.0.0.1:6379"
	}
	if conf.DB == 0 && db > 0 && db <= 15 {
		conf.DB = db
	}
	if conf.ConnTimeout == 0 {
		conf.ConnTimeout = 10000
	}
	if conf.ReadTimeout == 0 {
		conf.ReadTimeout = 10000
	}
	if conf.WriteTimeout == 0 {
		conf.WriteTimeout = 10000
	}
	if conf.LogSlow == 0 {
		conf.LogSlow = 100
	}

	redis := NewRedis(conf, nil)
	redisInstances[fullName] = redis
	return copyByLogger(redis, logger)
}

func copyByLogger(fromRedis *Redis, logger *log.Logger) *Redis {
	newRedis := new(Redis)
	newRedis.ReadTimeout = fromRedis.ReadTimeout
	newRedis.pool = fromRedis.pool
	newRedis.Config = fromRedis.Config
	if logger == nil {
		newRedis.logger = log.DefaultLogger
	} else {
		newRedis.logger = logger
	}
	return newRedis
}

func NewRedis(conf *Config, logger *log.Logger) *Redis {
	if logger == nil {
		logger = log.DefaultLogger
	}

	encryptedPassword := passwords[conf.Password]
	decryptedPassword := ""
	if encryptedPassword != "" {
		decryptedPassword = u.DecryptAes(encryptedPassword, settedKey, settedIv)
	} else {
		logger.Warning("Password is empty")
	}

	var redisReadTimeout time.Duration
	conn := &redis.Pool{
		MaxIdle:     conf.MaxIdles,
		MaxActive:   conf.MaxActive,
		IdleTimeout: time.Millisecond * time.Duration(conf.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			if conf.ReadTimeout > 0 {
				redisReadTimeout = time.Millisecond * time.Duration(conf.ReadTimeout)
			} else {
				redisReadTimeout = time.Millisecond * time.Duration(0)
			}
			c, err := redis.Dial("tcp", conf.Host,
				redis.DialConnectTimeout(time.Millisecond*time.Duration(conf.ConnTimeout)),
				redis.DialReadTimeout(redisReadTimeout),
				redis.DialWriteTimeout(time.Millisecond*time.Duration(conf.WriteTimeout)),
				redis.DialDatabase(conf.DB),
				redis.DialPassword(decryptedPassword),
			)
			if err != nil {
				logger.DBError(err.Error(), "redis", conf.Dsn(), "", nil, 0)
				return nil, err
			}
			//c.Do("SELECT", REDIS_DB)
			return c, nil
		},
	}

	redis := new(Redis)
	redis.ReadTimeout = conf.ReadTimeout
	redis.pool = conn
	redis.Config = conf
	if logger == nil {
		redis.logger = log.DefaultLogger
	} else {
		redis.logger = logger
	}

	return redis
}

func (rd *Redis) LogError(error string) {
	rd.logger.DBError(error, "redis", rd.Config.Dsn(), "", nil, 0)
}

func (rd *Redis) LogQuery(query string, args []interface{}, usedTime float32) {
	rd.logger.DB("redis", rd.Config.Dsn(), query, args, usedTime)
}

func (rd *Redis) LogQueryError(error string, query string, args []interface{}, usedTime float32) {
	rd.logger.DBError(error, "redis", rd.Config.Dsn(), query, args, usedTime)
}

func (rd *Redis) Destroy() error {
	if rd.pool == nil {
		return fmt.Errorf("operat on a bad redis pool")
	}
	err := rd.pool.Close()
	if err != nil {
		rd.LogError(err.Error())
	}
	return err
}

func (rd *Redis) GetPool() *redis.Pool {
	return rd.pool
}

func (rd *Redis) GetConnection() redis.Conn {
	if rd.pool == nil {
		return nil
	}
	return rd.pool.Get()
}

func (rd *Redis) Do(cmd string, values ...interface{}) *Result {
	if rd.pool == nil {
		err := errors.New("operat on a bad redis pool")
		rd.LogQueryError(err.Error(), cmd, values, 0)
		return &Result{Error: err}
	}
	startTime := time.Now()
	conn := rd.pool.Get()
	if conn.Err() != nil {
		rd.LogError(conn.Err().Error())
		return &Result{Error: conn.Err()}
	}
	r := rd.do(conn, cmd, values...)
	_ = conn.Close()
	usedTime := log.MakeUesdTime(startTime, time.Now())
	if r.Error == nil {
		if rd.Config.LogSlow == -1 || usedTime >= float32(rd.Config.LogSlow) {
			// 记录慢请求日志
			rd.LogQuery(cmd, values, usedTime)
		}
	} else {
		rd.LogQueryError(r.Error.Error(), cmd, values, usedTime)
	}
	return r
}

func (rd *Redis) do(conn redis.Conn, cmd string, values ...interface{}) *Result {
	cmdArr := strings.Split(cmd, " ")
	if len(cmdArr) > 1 {
		cmd = cmdArr[0]
		args := make([]interface{}, 0)
		for i := 1; i < len(cmdArr); i++ {
			args = append(args, cmdArr[i])
		}
		if len(values) > 0 {
			args = append(args, values...)
		}
		values = args
	}
	if strings.Contains(cmd, "MSET") {
		n := len(values)
		for i := n - 1; i > 0; i -= 2 {
			_checkValue(values, i)
		}
	} else if strings.Contains(cmd, "SET") {
		_checkValue(values, len(values)-1)
	}
	replyData, err := conn.Do(cmd, values...)
	if err != nil {
		return &Result{Error: err}
	}

	r := new(Result)
	switch realValue := replyData.(type) {
	case []byte:
		r.bytesData = realValue
	case string:
		r.bytesData = []byte(realValue)
	case int64:
		r.bytesData = []byte(strconv.FormatInt(realValue, 10))
	case []interface{}:
		if cmd == "HMGET" {
			r.keys = make([]string, len(values)-1)
			for i, v := range values {
				if i > 0 {
					r.keys[i-1] = u.String(v)
				}
			}
		} else if cmd == "MGET" {
			r.keys = make([]string, len(values))
			for i, v := range values {
				r.keys[i] = u.String(v)
			}
		}

		if cmd == "HGETALL" {
			r.keys = make([]string, len(realValue)/2)
			r.bytesDatas = make([][]byte, len(realValue)/2)
			i1 := 0
			i2 := 0
			for i, v := range realValue {
				if v != nil {
					if i%2 == 0 {
						r.keys[i1] = string(v.([]byte))
						i1++
					} else {
						switch subRealValue := v.(type) {
						case []byte:
							r.bytesDatas[i2] = subRealValue
						case string:
							r.bytesDatas[i2] = []byte(subRealValue)
						default:
							rd.LogError(fmt.Sprint("unknown reply type", cmd, i, v))
							r.bytesDatas[i2] = make([]byte, 0)
							r.Error = err
						}
						i2++
					}
				}
			}
		} else {
			r.bytesDatas = make([][]byte, len(realValue))
			for i, v := range realValue {
				if v != nil {
					switch subRealValue := v.(type) {
					case []byte:
						r.bytesDatas[i] = subRealValue
					case string:
						r.bytesDatas[i] = []byte(subRealValue)
					default:
						r.bytesDatas[i] = make([]byte, 0)
						r.Error = errors.New(fmt.Sprint("unknown reply type", cmd, i, v))
					}
				}
			}
		}
	case nil:
		r.bytesData = []byte{}
	default:
		err := fmt.Sprint("unknown reply type", cmd, reflect.TypeOf(replyData), replyData)
		r.Error = errors.New(err)
		r.bytesData = make([]byte, 0)
	}
	return r
}

func _checkValue(values []interface{}, index int) {
	if values[index] == nil {
		return
	}
	t := reflect.TypeOf(values[index])
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct || t.Kind() == reflect.Map || (t.Kind() == reflect.Slice && t.Elem().Kind() != reflect.Uint8) {
		encoded, err := json.Marshal(values[index])
		if err == nil {
			values[index] = encoded
		}
	}
}
