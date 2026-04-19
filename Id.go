// v4.3
package redis

import (
	"fmt"
	"sync"

	"github.com/ssgo/u"
)

type IdMaker struct {
	secCurrent   uint64
	secIndexNext uint64
	secIndexMax  uint64
	redisDownSec uint64
	secIndexLock sync.Mutex
	rd           *Redis
	maker        *u.IdMaker
}

func NewIdMaker(rd *Redis) *IdMaker {
	im := &IdMaker{rd: rd}
	im.maker = u.NewIdMaker(im.makeSecIndex)
	return im
}

func (im *IdMaker) makeSecIndex(sec uint64) uint64 {
	var secIndex uint64
	im.secIndexLock.Lock()
	defer im.secIndexLock.Unlock()
	if sec == im.redisDownSec {
		return 0
	}
	if im.secCurrent == sec && im.secIndexNext <= im.secIndexMax {
		secIndex = im.secIndexNext
		im.secIndexNext++
	} else {
		im.secCurrent = sec
		secIndexKey := fmt.Sprintf("_SecIdx_%d", sec)
		im.secIndexMax = uint64(im.rd.INCRBY(secIndexKey, 100))
		if im.secIndexMax >= 100 {
			secIndex = im.secIndexMax - 99
			im.secIndexNext = secIndex + 1
			if im.secIndexMax <= 100 {
				im.rd.EXPIRE(secIndexKey, 10)
			}
		} else {
			im.redisDownSec = sec
		}
	}
	return secIndex
}

func (im *IdMaker) Get(size int) string {
	return im.maker.Get(size)
}

func (im *IdMaker) GetForMysql(size int) string {
	return im.maker.GetForMysql(size)
}

func (im *IdMaker) GetForPostgreSQL(size int) string {
	return im.maker.GetForPostgreSQL(size)
}

// // v2.0
// var secCurrent uint64 = 0
// var secIndexNext uint64 = 0
// var secIndexMax uint64 = 0
// var secIndexLock = sync.Mutex{}

// // secTag（1位）+ sec（5位）+ secIndex（1+位），最小长度 7位
// // 142年后当 sec占 6位时通过 secTag（1位）区分防止碰撞，该算法可以确保8800年内不会发生碰撞
// // 8位：可以表示每5秒 3844个ID，每秒 768个
// // 10位：可以表示每5秒 1477万个ID，每秒 295万个
// // 12位：可以表示每5秒 568亿个ID，每秒 113亿个
// func (r *Redis) makeId(size int, ordered bool) string {
// 	tm := time.Now()
// 	// 计算从2000年开始到现在的索引值（2000年时间戳：946656000）（5位62进制最小值：14776336）（可以表示901356495个，即142.9年至2142年）
// 	sec := uint64((tm.Unix()-946656000)/5 + 14776336)

// 	secIndexKey := fmt.Sprintf("_SecIdx_%d", sec)
// 	var secIndex uint64
// 	secIndexLock.Lock()
// 	if secCurrent == sec && secIndexNext < secIndexMax {
// 		secIndex = secIndexNext
// 		secIndexNext++
// 	} else {
// 		secCurrent = sec
// 		secIndexMax = uint64(r.INCRBY(secIndexKey, 100))
// 		if secIndexMax == 0 {
// 			secIndexMax = u.GlobalRand2.Uint64N(14776236) + 100
// 		}
// 		secIndex = secIndexMax - 100
// 		secIndexNext = secIndex + 1
// 		if secIndexMax <= 100 {
// 			r.EXPIRE(secIndexKey, 10)
// 		}
// 	}
// 	secIndexLock.Unlock()

// 	// 计算 secTag（通过secIndexLen来防止不同长度的secIndex导致碰撞），142年内通过 tm.UnixMicro()%7 产生随机性，142年后通过放弃随机性使用固定的49来跟前142年区分开
// 	intEncoder := u.DefaultIntEncoder
// 	if ordered {
// 		intEncoder = u.OrderedIntEncoder
// 	}
// 	secBytes := intEncoder.EncodeInt(sec)
// 	secLen := len(secBytes)
// 	inSecIndexBytes := intEncoder.EncodeInt(secIndex)
// 	secIndexLen := uint64(len(inSecIndexBytes))
// 	var uid = make([]byte, 0, size)
// 	if secLen <= 5 {
// 		uid = intEncoder.AppendInt(uid, uint64(tm.UnixMicro()%7)*7+secIndexLen)
// 	} else {
// 		uid = intEncoder.AppendInt(uid, 49+secIndexLen)
// 	}

// 	uid = append(uid, secBytes...)
// 	uid = append(uid, inSecIndexBytes...)
// 	uid = intEncoder.FillInt(uid, size) // 用随机数填充
// 	if !ordered {
// 		uid = intEncoder.HashInt(u.ExchangeInt(uid)) // 交叉然后散列乱序
// 	} else {
// 		postBytes := intEncoder.HashInt(u.ExchangeInt(uid[secLen+1:])) // 交叉然后散列乱序
// 		copy(uid[secLen+1:], postBytes)
// 	}
// 	return string(uid)
// }

// func (r *Redis) MakeId(size int) string {
// 	return r.makeId(size, false)
// }

// func (r *Redis) MakeOrderedId(size int) string {
// 	return r.makeId(size, true)
// }
