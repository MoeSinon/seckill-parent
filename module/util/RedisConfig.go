package util

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
)

func DoJsonEncodingStore(testStruct struct{}, conn redis.Conn) {
	//json序列化
	datas, _ := json.Marshal(testStruct)
	//缓存数据
	conn.Do("set", "struct3", datas)
}
func DoJsonDEcodingStore(testStruct struct{}, conn redis.Conn) {
	//读取数据
	rebytes, _ := redis.Bytes(conn.Do("get", "struct3"))
	//json反序列化
	object := &testStruct
	json.Unmarshal(rebytes, object)
}
