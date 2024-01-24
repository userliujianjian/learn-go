package main

type Once chan struct{}

func NewOnce() Once {
	o := make(Once, 1)
	o <- struct{}{}
	return o
}

func (o Once) Do(f func()) {
	_, ok := <-o
	if !ok {
		return
	}

	f()

	close(o)
}

type Semaphore chan struct{}

func NewSemaphore(size int) Semaphore {
	return make(Semaphore, size)
}

func (s Semaphore) Lock() {
	// Writes will only succeed if there is room in s.
	s <- struct{}{}
}

// TryLock is like Lock but it immediately returns whether it was able
// to lock or not without waiting.
func (s Semaphore) TryLock() bool {
	// Select with default case: if no cases ar ready
	// just fall in the default block
	select {
	case s <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s Semaphore) Unlock() {
	// Make room for other users of the semaphore
	<-s
}

type Mutex Semaphore

func NewMutex() Mutex {
	return Mutex(NewSemaphore(1))
}
