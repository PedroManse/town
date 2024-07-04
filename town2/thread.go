package town2

import (
	"sync"
)

type Thread[I, O any] struct {
	state    ThreadState
	flags    ThreadFlag
	inchan   chan I
	outchan  chan O
	diechan  chan struct{}
	iolock   sync.Mutex
	proclock sync.Mutex
}

func NewThread[I, O any](
	fn func(*Thread[I, O]),
	inlen int,
	outlen int,
) *Thread[I, O] {
	inchan := make(chan I, inlen)
	outchan := make(chan O, outlen)
	diechan := make(chan struct{})

	thr := Thread[I, O]{
		state:    userOwned,
		flags:    running,
		inchan:   inchan,
		outchan:  outchan,
		diechan:  diechan,
		iolock:   sync.Mutex{},
		proclock: sync.Mutex{},
	}

	go fn(&thr)
	return &thr
}

// impl Pipe
func (t *Thread[I, O]) Send(i I) {
	t.iolock.Lock()
	t.iolock.Unlock()
	t.inchan <- i
}
func (t *Thread[I, O]) Await() O {
	t.iolock.Lock()
	t.iolock.Unlock()
	return <-t.outchan
}
func (t *Thread[I, O]) Die() {
	t.diechan <- struct{}{}
}
func (t *Thread[I, O]) InQueue() int {
	return len(t.inchan)
}
func (t *Thread[I, O]) InMaxQueue() int {
	return cap(t.inchan)
}
func (t *Thread[I, O]) OutQueue() int {
	return len(t.outchan)
}
func (t *Thread[I, O]) OutMaxQueue() int {
	return cap(t.outchan)
}
func (t *Thread[I, O]) LockIO() (chan I, chan O) {
	t.iolock.Lock()
	t.state = locked
	return t.inchan, t.outchan
}
func (t *Thread[I, O]) Own() (chan I, chan O) {
	t.state = threadOwned
	return t.inchan, t.outchan
}
func (t *Thread[I, O]) LockProc() {
	t.proclock.Lock()
}
func (t *Thread[I, O]) UnlockProc() {
	t.proclock.Unlock()
}
func (t *Thread[I, O]) UnlockIO(chan I, chan O) {
	t.state = userOwned
	t.iolock.Unlock()
}
