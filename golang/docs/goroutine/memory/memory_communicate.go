package memory

import (
	"sync"
	"time"
)

type Resource struct {
	url        string
	polling    bool
	lastPolled int64
}

type Resources struct {
	data []*Resource
	lock *sync.Mutex
}

func Poller(res *Resources) {
	for {
		//	 get the least recently-polled Resource
		// and mark it as being polled
		res.lock.Lock()
		var r *Resource
		for _, v := range res.data {
			if v.polling {
				continue
			}

			if r == nil || v.lastPolled < r.lastPolled {
				r = v
			}

		}

		if r != nil {
			r.polling = true
		}
		res.lock.Unlock()

		if r == nil {
			continue
		}

		// pool the url

		// update the resource's polling and lastPolled
		res.lock.Lock()
		r.polling = false
		r.lastPolled = int64(time.Nanosecond)
		res.lock.Unlock()
	}
}

type Source string

func Poller2(in, out chan *Source) {
	for r := range in {
		// poll the URL

		// send the processed Resource to out
		out <- r
	}
}
