package main

import (
	"fmt"
	"math/rand"
	"time"
)

// RSS 客户端运行
// Item 一个RSS元素
type Item struct {
	Title, Channel, GUID string
}

// Fetcher Fetcher获取Items并返回下一次获取时间
// 失败时将返回非nil错误
type Fetcher interface {
	Fetch() (items []Item, next time.Time, err error)
}

// Subscription 订阅通过渠道传送项目。Close取消subscription，关闭Updates通道，并返回上次提取错误。
type Subscription interface {
	Updates() <-chan Item
	Close() error
}

func Subscribe(fetcher Fetcher) Subscription {
	s := &sub{
		fetcher: fetcher,
		updates: make(chan Item),
		closing: make(chan chan error),
	}
	go s.loop()
	return s
}

func NaiveMerge(subs ...Subscription) Subscription {
	m := &naiveMerge{
		subs:    subs,
		updates: make(chan Item),
	}

	for _, sb := range subs {
		go func(s Subscription) {
			for it := range s.Updates() {
				m.updates <- it
			}
		}(sb)
	}
	return m
}

func Merge(subs ...Subscription) Subscription {
	m := &merge{
		subs:    subs,
		updates: make(chan Item),
		quit:    make(chan struct{}),
		errs:    make(chan error),
	}

	for _, sb := range subs {
		go func(s Subscription) {
			var it Item
			select {
			case it = <-s.Updates():
			case <-m.quit:
				m.errs <- s.Close()
				return
			}
			select {
			case m.updates <- it:
			case <-m.quit:
				m.errs <- s.Close()
				return
			}
		}(sb)
	}
	return m
}

func Fetch(domain string) Fetcher {
	return fakeFetch(domain)
}

func fakeFetch(domain string) Fetcher {
	return &fakeFetcher{channel: domain}
}

func NaiveSubscribe(fetcher Fetcher) Subscription {
	s := &naiveSub{
		fetcher: fetcher,
		updates: make(chan Item),
	}
	go s.loop()
	return s
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	merged := Merge(
		NaiveSubscribe(Fetch("blog.golang.org")),
		NaiveSubscribe(Fetch("googleblog.blogspot.com")),
		NaiveSubscribe(Fetch("googledevelopers.blogspot.com")),
	)
	time.AfterFunc(3*time.Second, func() {
		fmt.Println("closed: ", merged.Close())
	})

	for it := range merged.Updates() {
		fmt.Println(it.Channel, it.Title)
	}

	time.Sleep(10 * time.Second)

	panic("show me the stacks")
}
