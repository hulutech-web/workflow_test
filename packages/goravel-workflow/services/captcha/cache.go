package captcha

import (
	"sync"
	"time"
)

// cachedata 缓存数据结构，包含缓存数据本身及其创建时间。
type cachedata = struct {
	data     []byte
	createAt time.Time
}

// mux 互斥锁，用于保证对 cachemaps 的并发读写安全。
var mux sync.Mutex

// cachemaps 全局缓存映射表，以字符串 key 索引缓存数据。
var cachemaps = make(map[string]*cachedata)

// WriteCache 将验证码数据写入缓存。
// key: 缓存键（通常为验证码 ID）。
// data: 待缓存的验证码图片字节数据。
func WriteCache(key string, data []byte) {
	mux.Lock()
	defer mux.Unlock()
	cachemaps[key] = &cachedata{
		createAt: time.Now(),
		data:     data,
	}
}

// ReadCache 从缓存中读取验证码数据。
// key: 缓存键。
// 返回缓存中的字节数据；如果 key 不存在，则返回空字节切片。
func ReadCache(key string) []byte {
	mux.Lock()
	defer mux.Unlock()
	if cd, ok := cachemaps[key]; ok {
		return cd.data
	}

	return []byte{}
}

// ClearCache 从缓存中删除指定 key 的缓存条目。
// key: 待删除的缓存键。
func ClearCache(key string) {
	mux.Lock()
	defer mux.Unlock()
	delete(cachemaps, key)
}

// RunTimedTask 启动定时清理任务，每 5 分钟检查并清除过期的缓存条目。
// 该函数启动一个后台 goroutine，在应用运行期间持续执行。
func RunTimedTask() {
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for range ticker.C {
			checkCacheOvertimeFile()
		}
	}()
}

// checkCacheOvertimeFile 遍历所有缓存条目，删除创建时间超过 30 分钟的过期数据。
func checkCacheOvertimeFile() {
	// 收集所有过期的 key
	var keys = make([]string, 0)
	for key, data := range cachemaps {
		ex := time.Now().Unix() - data.createAt.Unix()
		if ex > (60 * 30) {
			keys = append(keys, key)
		}
	}

	// 批量删除过期的缓存条目
	for _, key := range keys {
		ClearCache(key)
	}
}
