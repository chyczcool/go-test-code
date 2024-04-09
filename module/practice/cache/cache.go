/*
1、支持设置过期时间，支持到秒
2、支持设置最大内存，当内存超出时做出合适的处理
3、支持并发安全
4、接口实现要求

	type Cache interface {
		//size: 1KB 100KB 1MB 2MB 1GB
		SetMaxMemory(size string) bool
		//将Value写入缓存
		Set(key string, val any, expire time.Duration) bool
		//根据key值获取value
		Get(key string) (any, bool)
		//删除key值
		Del(key string) bool
		//判断key值是否存在
		Exists(key string) bool
		//清空所有key
		Flush() bool
		//获取缓存中所有key的数量
		Keys() int64
	}

5、使用示例
cache := NewMemCache()
cache.SetMaxMemory("100MB")
cache.Set("int", 1)
cache.Set("bool", false)
cache.Set("data", map[string]any{"a":1})
cache.Get("int")
cache.Del("int")
cache.Flush()
cache.Keys()
*/
package cache

import (
	"strconv"
	"strings"
	"time"
)

const (
	MAX_MEM_SIZE = 4 * 1024 * 1024 * 1024
)

type Cache interface {
	//size: 1KB 100KB 1MB 2MB 1GB
	SetMaxMemory(size string) bool
	//将Value写入缓存
	Set(key string, val any, expire time.Duration) bool
	//根据key值获取value
	Get(key string) (any, bool)
	//删除key值
	Del(key string) bool
	//判断key值是否存在
	Exists(key string) bool
	//清空所有key
	Flush() bool
	//获取缓存中所有key的数量
	Keys() int64
}

type MemAche struct {
	Cache

	memCache map[string]any
	nowBytes int64
	maxBytes int64
}

func NewMemCache() *MemAche {
	return &MemAche{memCache: make(map[string]any)}
}

func checkMemSize(size string) (num int64, ok bool) {
	if num, err := strconv.ParseInt(size, 10, 64); err != nil {
		println("memery unit error. %s", err)
		return 0, false
	} else {
		return num, true
	}

}

// 设置缓存大小
func (cache *MemAche) SetMaxMemory(size string) bool {

	var bytesSize int64
	upperSize := strings.ToUpper(size)
	switch {
	case strings.HasSuffix(upperSize, "KB"):
		num, ok := checkMemSize(upperSize[:len(upperSize)-2])
		if !ok {
			return false
		}
		bytesSize = num * 1024
	case strings.HasSuffix(upperSize, "MB"):
		num, ok := checkMemSize(upperSize[:len(upperSize)-2])
		if !ok {
			return false
		}
		bytesSize = num * 1024 * 1024
	case strings.HasSuffix(upperSize, "GB"):
		num, ok := checkMemSize(upperSize[:len(upperSize)-2])
		if !ok {
			return false
		}
		bytesSize = num * 1024 * 1024 * 1024
	}

	if bytesSize > MAX_MEM_SIZE {
		println("Too big, please lower than 4G.")
		return false
	}
	//TODO 判断是否小于当前内存大小

	cache.maxBytes = bytesSize
	return true
}

// 将Value写入缓存
func (cache *MemAche) Set(key string, val any, expire time.Duration) bool {
	switch val.

	if cache.nowBytes +  > cache.maxBytes {

	}

	return true
}

// 根据key值获取value
func (cache *MemAche) Get(key string) (any, bool) {
	return 0, true
}

// 删除key值
func (cache *MemAche) Del(key string) bool

// 判断key值是否存在
func (cache *MemAche) Exists(key string) bool

// 清空所有key
func (cache *MemAche) Flush() bool

// 获取缓存中所有key的数量
func (cache *MemAche) Keys() int64
