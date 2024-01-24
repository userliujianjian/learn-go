package main

type RWMutex struct {
	write   chan struct{}
	readers chan int
}

func NewLock() RWMutex {
	return RWMutex{
		// This is used as a normal mutex.
		write: make(chan struct{}, 1),
		// This is used to protect the readers count.
		// By receiving the value it is guaranteed that no
		// other goroutine is changing it at the same time.
		readers: make(chan int, 1),
	}
}

func (l RWMutex) Lock() {
	l.write <- struct{}{}
}

func (l RWMutex) Unlock() {
	<-l.write
}

func (l RWMutex) RLock() {
	// Count current readers. Default to 0.
	var rs int
	// Select on the channels without default.
	// One and only one case will be selected and this
	// will block until one case becomes available.
	select {
	case l.write <- struct{}{}: // One sending case for write.
		// If the write lock is available we have no readers.
		// we grab the write lock to prevent concurrent
		// read-writes.
	case rs = <-l.readers:
		//	There already ar readers, let's grab and update the
		// readers count.

	}

	rs++

	l.readers <- rs
}

func (l RWMutex) RUnlock() {
	rs := <-l.readers
	rs--
	if rs == 0 {
		<-l.write
		return
	}

	l.readers <- rs
}

func (l RWMutex) TryLock() bool {
	select {
	case l.write <- struct{}{}:
		return true
	default:
		return false
	}
}

func (l RWMutex) TryRLock() bool {
	var rs int
	select {
	case l.write <- struct{}{}:
	case rs = <-l.readers:
	default:
		return false

	}
	rs++
	l.readers <- rs
	return true
}
