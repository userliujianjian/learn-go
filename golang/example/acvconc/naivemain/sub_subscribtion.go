package main

import "time"

type sub struct {
	fetcher Fetcher
	updates chan Item
	closing chan chan error // for close
}

func (s *sub) Updates() <-chan Item {
	return s.updates
}

func (s *sub) Close() error {
	errc := make(chan error)
	s.closing <- errc
	return <-errc
}

func (s *sub) loopCloseOnly() {
	var err error
	for {
		select {
		case errc := <-s.closing:
			errc <- err
			close(s.updates)
			return
		}
	}
}

func (s *sub) LoopFetchOnly() {
	var pending []Item
	var next time.Time
	var err error

	for {
		var fetchDelay time.Duration
		if now := time.Now(); next.After(now) {
			fetchDelay = next.Sub(now)
		}
		startFetch := time.After(fetchDelay)

		select {
		case <-startFetch:
			var fetched []Item
			fetched, next, err = s.fetcher.Fetch()
			if err != nil {
				next = time.Now().Add(10 * time.Second)
				break
			}
			pending = append(pending, fetched...)
		}
	}

	//.....
}

func (s *sub) loopSendOnly() {
	var pending []Item
	for {
		var first Item
		var updates chan Item
		if len(pending) > 0 {
			first = pending[0]
			updates = s.updates
		}

		select {
		case updates <- first:
			pending = pending[1:]
		}
	}

	// stop send only omit
}

func (s *sub) mergedLoop() {
	var pending []Item
	var next time.Time
	var err error

	for {
		var fetchDelay time.Duration
		if now := time.Now(); next.After(now) {
			fetchDelay = next.Sub(now)
		}

		startFetch := time.After(fetchDelay)
		var first Item
		var updates chan Item
		if len(pending) > 0 {
			first = pending[0]
			updates = s.updates
		}

		select {
		case errc := <-s.closing:
			errc <- err
			close(s.updates)
			return
		case <-startFetch:
			var fetched []Item
			fetched, next, err = s.fetcher.Fetch()
			if err != nil {
				next = time.Now().Add(10 * time.Second)
				break
			}
			pending = append(pending, fetched...)
		case updates <- first:
			pending = pending[1:]
		}

	}
}

func (s *sub) dedupeLoop() {
	const maxPending = 10

	var pending []Item
	var next time.Time
	var err error
	var seen = make(map[string]bool) // set of item.GUIDS

	for {
		var fetchDelay time.Duration
		if now := time.Now(); next.After(now) {
			fetchDelay = next.Sub(now)
		}

		var startFetch <-chan time.Time
		if len(pending) < maxPending {
			startFetch = time.After(fetchDelay)
		}

		var first Item
		var updates chan Item
		if len(pending) > 0 {
			first = pending[0]
			updates = s.updates
		}

		select {
		case errc := <-s.closing:
			errc <- err
			close(s.updates)
			return
		case <-startFetch:
			var fetched []Item
			fetched, next, err = s.fetcher.Fetch()
			if err != nil {
				next = time.Now().Add(10 * time.Second)
				break
			}
			for _, item := range fetched {
				if !seen[item.GUID] {
					pending = append(pending, item)
					seen[item.GUID] = true
				}
			}
		case updates <- first:
			pending = pending[1:]

		}
	}
}

func (s *sub) loop() {
	const maxPending = 10
	type fetchResult struct {
		fetched []Item
		next    time.Time
		err     error
	}

	var fetchDone chan fetchResult
	var pending []Item
	var next time.Time
	var err error
	var seen = make(map[string]bool)

	for {
		var fetchDelay time.Duration
		if now := time.Now(); next.After(now) {
			fetchDelay = next.Sub(now)
		}
		var startFetch <-chan time.Time
		if fetchDone == nil && len(pending) < maxPending {
			startFetch = time.After(fetchDelay)
		}

		var first Item
		var updates chan Item
		if len(pending) > 0 {
			first = pending[0]
			updates = s.updates
		}

		select {
		case <-startFetch:
			fetchDone = make(chan fetchResult, 1)
			go func() {
				fetched, next, err := s.fetcher.Fetch()
				fetchDone <- fetchResult{fetched, next, err}
			}()
		case result := <-fetchDone:
			fetchDone = nil
			fetched := result.fetched
			next, err = result.next, result.err
			if err != nil {
				next = time.Now().Add(10 * time.Second)
				break
			}
			for _, item := range fetched {
				if id := item.GUID; !seen[id] {
					pending = append(pending, item)
					seen[id] = true
				}
			}
		case errc := <-s.closing:
			errc <- err
			close(s.updates)
			return
		case updates <- first:
			pending = pending[1:]

		}

	}
}
