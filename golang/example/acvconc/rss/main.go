package main

import "time"

type Item struct {
	Title, channel, Guid string // a subset of RSS fields
}

// bug i want a stream:
// <- chan Item

type Fetcher interface {
	Fetch() (items []Item, next time.Time, err error)
}

//func Fetch(domain string) Fetcher{} // fetchs items for domain

type Subscription interface {
	Updates() <-chan Item // stream of Items
	Close() error         // shuts down the stream
}

//func Subscribe(fetcher Fetcher) Subscription{...} // 将Fetches转换为流

//func Merge(subs ...Subscription) Subscription{...} // 合并多个流

type sub struct {
	fetcher Fetcher   // fetch items
	updates chan Item // delivers items to the user
}

func (s *sub) loop() {
	// TODO
}

func (s *sub) Updates() <-chan Item {
	return s.updates
}

func (s *sub) Close() error {
	// TODO: make loop exit
	// TODO: find out about any error
	return nil
}

func Subscribe(fetcher Fetcher) Subscription {
	s := &sub{
		fetcher: fetcher,
		updates: make(chan Item),
	}

	go s.loop()
	return s
}

func (s *sub) test() {

}
