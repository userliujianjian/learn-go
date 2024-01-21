package main

type deduper struct {
	s       Subscription
	updates chan Item
	closing chan chan error
}

func (d *deduper) loop() {
	in := d.s.Updates()
	var pending Item
	var out chan Item
	seen := make(map[string]bool)

	for {
		select {
		case it := <-in:
			if !seen[it.GUID] {
				pending = it
				in = nil
				out = d.updates
				seen[it.GUID] = true
			}
		case out <- pending:
			in = d.s.Updates()
			out = nil
		}
	}
}

func (d *deduper) Close() error {
	errc := make(chan error)
	d.closing <- errc
	return <-errc
}

func (d *deduper) Updates() <-chan Item {
	return d.updates
}
