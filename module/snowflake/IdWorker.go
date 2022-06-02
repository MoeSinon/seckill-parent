package snowflake

import (
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"
)

/**
 * 核心代码为其Idslaver这个类实现，其原理结构如下，我分别用一个0表示一位，用—分割开部分的作用：
 * 1||0---0000000000 0000000000 0000000000 0000000000 0 --- 0000000 ---000 ---000000000000
 * 在上面的字符串中，第一位为未使用（实际上也可作为long的符号位），接下来的41位为毫秒级时间，
 * 然后5位master标识位，5位机器ID（并不算标识符，实际是为线程标识），
 * 然后12位该毫秒内的当前毫秒内的计数，加起来刚好64位。
 * 这样的好处是，整体上按照时间自增排序，并且整个分布式系统内不会产生ID碰撞（由master和机器ID作区分），
 * 并且效率较高，经测试，snowflake每秒能够产生26万ID左右，完全满足需要。
 * <p>
 * 64位ID (42(毫秒)+5(机器ID)+5(业务编码)+12(重复累加))
 *
 * @author Polim
 */

type Snowflake struct {
	sync.Mutex       // 锁
	timestamp  int64 // 时间戳 ，毫秒
	slaverId   int64 // 工作节点
	masterId   int64 // 数据中心机房id
	sequence   int64 // 序列号
}

const (
	// 时间起始标记点，作为基准，一般取系统的最近时间（一旦确定不能变动）
	twipoch       int64 = 1652332995000
	timestampBits uint8 = 41
	// 机器标识位数
	slaverIdBits uint8 = 3
	// 数据中心标识位数
	masterIdBits uint8 = 7
	// 毫秒内自增位
	sequenceBits uint8 = 12
	// 机器ID最大值
	maxslaverId int64 = -1 ^ (-1 << slaverIdBits)
	// 数据中心ID最大值
	maxmasterId  int64 = -1 ^ (-1 << masterIdBits)
	sequenceMask int64 = -1 ^ (-1 << sequenceBits)
	// 机器ID偏左移12位
	slaverIdShift uint8 = sequenceBits
	// 数据中心ID左移17位
	masterIdShift uint8 = sequenceBits + slaverIdBits
	// 时间毫秒左移22位
	timestampLeftShift uint8 = sequenceBits + slaverIdBits + masterIdBits
	/* 上次生产id时间戳 */
	lastTimestamp int64 = -1 ^ (-1 << timestampBits)
	// 0，并发控制
	sequence int = 0
)

func (s *Snowflake) GetId() (int64, int64) {
	var ipint, sId int
	network, _ := net.InterfaceAddrs()
	if len(network) == 0 {
		s.masterId = 1
		sId += 1
		s.masterId = int64(sId)
	} else {
		for _, netInterface := range network {
			if ip, ok := netInterface.(*net.IPNet); ok && !ip.IP.IsLoopback() {
				ipint := big.NewInt(0).SetBytes([]byte(netInterface.String())).Int64()
				s.masterId = int64(ipint)
			} else {
				s.masterId = 1
				sId += 1
			}
		}
	}
	// rand.Seed(time.Now().UnixNano())
	// s.slaverId = rand.Int63n(int64(masterIdBits))
	s.slaverId = int64(sId)
	return int64(ipint), s.slaverId
}

func retureSnowflake() *Snowflake {
	return &Snowflake{}
}

func (s *Snowflake) NextVal() int64 {
	s.Lock()
	now := time.Now().UnixNano() / 1000000 // 转毫秒
	if s.timestamp == now {
		// 当同一时间戳（精度：毫秒）下多次生成id会增加序列号
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 如果当前序列超出12bit长度，则需要等待下一毫秒
			// 下一毫秒将使用sequence:0
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		// 不同时间戳（精度：毫秒）下直接使用序列号：0
		s.sequence = 0
	}

	t := now - twipoch
	if t > lastTimestamp {
		s.Unlock()
		fmt.Printf("epoch must be between 0 and %d", lastTimestamp-1)
		return 0
	}
	s.timestamp = now
	s.GetId()
	r := int64((s.timestamp << timestampLeftShift) | (s.masterId << masterIdShift) | (s.slaverId << slaverIdShift) | (s.sequence))
	s.Unlock()
	return r
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			var uuids sync.Map
			// uuids := make(map[string]bool)
			snowflakes := retureSnowflake()
			for i := 0; i < 100; i++ {
				uuid := fmt.Sprintf("%b\n", snowflakes.NextVal())
				if _, ok := uuids.Load(uuid); ok {
					fmt.Println(uuid)
				}
				uuids.Store(uuid, true)
				// if uuids[uuid] {
				// 	fmt.Println(uuid)
				// }
			}

		}()
		wg.Done()
	}
	wg.Wait()
}
