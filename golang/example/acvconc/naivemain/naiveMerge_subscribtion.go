package main

type naiveMerge struct {
	subs    []Subscription
	updates chan Item
}

func (m *naiveMerge) Close() (err error) {
	for _, sub := range m.subs {
		if e := sub.Close(); err == nil && e != nil {
			err = e
		}
	}
	close(m.updates)
	return
}

func (m *naiveMerge) Updates() <-chan Item {
	return m.updates
}
