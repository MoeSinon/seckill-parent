package Snowflake

import (
	"sync"
	"time"
)

/**
 * <p>描述：分布式自增长ID</p>
 * <pre>
 *     Twitter的 Snowflake　JAVA实现方案
 * </pre>
 * 核心代码为其Idslaver这个类实现，其原理结构如下，我分别用一个0表示一位，用—分割开部分的作用：
 * 1||0---0000000000 0000000000 0000000000 0000000000 0 --- 00000 ---00000 ---000000000000
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
	sync.Mutex         // 锁
	timestamp    int64 // 时间戳 ，毫秒
	workerid     int64 // 工作节点
	datacenterid int64 // 数据中心机房id
	sequence     int64 // 序列号
}

const (
	// 时间起始标记点，作为基准，一般取系统的最近时间（一旦确定不能变动）
	twipoch int64 = 1652332995000
	// 机器标识位数
	slaverIdBits uint8 = 5
	// 数据中心标识位数
	masterIdBits uint8 = 5
	// 机器ID最大值
	maxslaverId int64 = -1 ^ (-1 << slaverIdBits)
	// 数据中心ID最大值
	maxmasterId int64 = -1 ^ (-1 << masterIdBits)
	// 毫秒内自增位
	sequenceBits uint8 = 12
	// 机器ID偏左移12位
	slaverIdShift uint8 = sequenceBits
	// 数据中心ID左移17位
	masterIdShift uint8 = sequenceBits + slaverIdBits
	// 时间毫秒左移22位
	timestampLeftShift uint8 = sequenceBits + slaverIdBits + masterIdBits

	sequenceMask int64 = -1 ^ (-1 << sequenceBits)
	/* 上次生产id时间戳 */
	lastTimestamp uint8 = -1
	// 0，并发控制
	sequence int = 0
)

// func Idslaver(){
//     masterId = getmasterId(maxmasterId)
//     slaverId = getMaxslaverId(masterId, maxslaverId)

//     /**
//      * @param slaverId
//      *            工作机器ID
//      * @param masterId
//      *            序列号
//      */
// func Idslaver(long slaverId, long masterId) {
//     if (slaverId > maxslaverId || slaverId < 0) {
//         throw new IllegalArgumentException(String.format("slaver Id can't be greater than %d or less than 0", maxslaverId))
//     }
//     if (masterId > maxmasterId || masterId < 0) {
//         throw new IllegalArgumentException(String.format("master Id can't be greater than %d or less than 0", maxmasterId))
//     }
//     slaverId = slaverId
//     masterId = masterId
// }
//     /**
//      * 获取下一个ID
//      *
//      * @return
//      */
//     func synchronized long nextId() {
//         long timestamp = timeGen()
//         if (timestamp < lastTimestamp) {
//             throw new RuntimeException(String.format("Clock moved backwards.  Refusing to generate id for %d milliseconds", lastTimestamp - timestamp))
//         }

//         if (lastTimestamp == timestamp) {
//             // 当前毫秒内，则+1
//             sequence = (sequence + 1) & sequenceMask
//             if (sequence == 0) {
//                 // 当前毫秒内计数满了，则等待下一秒
//                 timestamp = tilNextMillis(lastTimestamp)
//             }
//         } else {
//             sequence = 0L
//         }
//         lastTimestamp = timestamp
//         // ID偏移组合生成最终的ID，并返回ID
//         long nextId = ((timestamp - twepoch) << timestampLeftShift)
//                 | (masterId << masterIdShift)
//                 | (slaverId << slaverIdShift) | sequence

//         return nextId
//     }

//     tilNextMillis(final long lastTimestamp) {
//         long timestamp = timeGen()
//         while (timestamp <= lastTimestamp) {
//             timestamp = timeGen()
//         }
//         return timestamp
//     }

//     timeGen() {
//         return System.currentTimeMillis()
//     }

//     /**
//      * <p>
//      * 获取 maxslaverId
//      * </p>
//      */
// func getMaxslaverId() {
//     var mpid []byte
//     mpid.append(mpid,masterId)
//     String name = ManagementFactory.getRuntimeMXBean().getName()
//     if (!name.isEmpty()) {
//         /*
//             * GET jvmPid
//             */
//         mpid.append(name.split("@")[0])
//     }
//     /*
//         * MAC + PID 的 hashcode 获取16个低位
//         */
//     return (mpid.toString().hashCode() & 0xffff) % (maxslaverId + 1)
// }

//     /**
//      * <p>
//      * 数据标识id部分
//      * </p>
//      */
// func getmasterId() unint8 {
// 	id := 0
// 	// InetAddress ip = InetAddress.getLocalHost()
// 	network, _ := net.InterfaceAddrs()
// 	if len(network) == 0 {
// 		id = 1
// 	} else {
// 		for _, netInterface := range network {
//             if mac != net.FlagLoopback.String() {
//                 mac := netInterface.String()
//             }
// 			id = ((0x000000FF && mac[:len(mac) - 1]) || (0x0000FF00 && mac[:len(mac) - 2]) << 8) >> 6

// 			id = id % (maxmasterId + 1)
// 		}

// 		    catch (Exception e) {
// 		        System.out.println(" getmasterId: " + e.getMessage())
// 	}
// 	    return id
// }
// }

// func main(String[] args) {
//     //推特->雪花算法->只要时间不倒流，算出的ID永不冲突   每秒能产生26万个ID
//     Idslaver idslaver=new Idslaver(0,0)
//     for(int i=0i<1000i++){
//         long nextId = idslaver.nextId()
//         System.out.println(nextId)
//     }
// }
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
	t := now - epoch
	if t > timestampMax {
		s.Unlock()
		glog.Errorf("epoch must be between 0 and %d", timestampMax-1)
		return 0
	}
	s.timestamp = now
	r := int64((t)<<timestampShift | (s.datacenterid << datacenteridShift) | (s.workerid << workeridShift) | (s.sequence))
	s.Unlock()
	return r
}
