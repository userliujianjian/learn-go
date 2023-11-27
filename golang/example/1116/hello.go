package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 结构图--中间件
type APIKey struct {
	ID        string
	UpdatedAt time.Time
}

// 模拟中间件抓去数据
func getAPIKey(r *http.Request) (k APIKey, err error) {
	// 通过上下文拿到数据并存入队列中

	// 定义一个队列

	return
}

// 保存数据库 --- demo
func saveToDB(data []*ApiDbData) (err error) {

	return nil
}

type ApiDbData struct {
	UserId   string
	ApiKey   string
	LastTime time.Time
}

type AppData struct {
	UserDataMap          map[string]map[string]*APIKey
	Queue                []*ApiDbData
	Mutex                *sync.Mutex
	LastQueueProcessTime time.Time
}

func Middleware(appData *AppData, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取用户标识
		userId := r.Header.Get("User-Id")

		// 保证并发
		appData.Mutex.Lock()
		defer appData.Mutex.Unlock()

		// 获取apikey
		key, err := getAPIKey(r)
		if err != nil {
			return
		}

		//检查map中是否存在
		userMap, ok := appData.UserDataMap[userId]
		if !ok {
			userMap = make(map[string]*APIKey)
			appData.UserDataMap[userId] = userMap
		}
		v := &ApiDbData{
			ApiKey:   key.ID,
			UserId:   userId,
			LastTime: key.UpdatedAt,
		}

		appData.Queue = append(appData.Queue, v)
		// 检测队列条数是否大于100，或者是否已经30s没有更新
		if len(appData.Queue) >= 100 || time.Since(appData.LastQueueProcessTime) >= 30 {
			go func() {
				defer func() {
					if err2 := recover(); err2 != nil {
						fmt.Println("Recover from panic: ", err2)
					}
				}()

				err = saveToDB(appData.Queue)
				if err != nil {
					// TODO
					return
				}

				appData.Queue = nil
				appData.LastQueueProcessTime = time.Now()
			}()
		}
		// 执行下一个处理器
		next.ServeHTTP(w, r)

	})
}

// 入口
func main() {
	appData := &AppData{
		LastQueueProcessTime: time.Now(),
		UserDataMap:          make(map[string]map[string]*APIKey),
		Queue:                nil,
	}
	// TODO 这里到底是什么原理呢？
	hello := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world")
	})
	http.Handle("/", Middleware(appData, hello))
	http.ListenAndServe(":8801", nil)
}

/// 没有阻塞程序结束运行

// 创建一个简单的http服务
//
