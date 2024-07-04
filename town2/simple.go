package town2

import (
	"sync"
)

type SimpleThread[I, O any] struct {
	state    ThreadState
	flags    ThreadFlag
	fn       func(I) O
	inchan   chan I
	outchan  chan O
	diechan  chan struct{}
	iolock   sync.Mutex
	proclock sync.Mutex
}

func ContainSimple[I, O any](
	fn func(I) O,
	inchan chan I,
	outchan chan O,
) *SimpleThread[I, O] {
	diechan := make(chan struct{})

	thr := SimpleThread[I, O]{
		state:    userOwned,
		flags:    simple + running,
		fn:       fn,
		inchan:   inchan,
		outchan:  outchan,
		diechan:  diechan,
		iolock:   sync.Mutex{},
		proclock: sync.Mutex{},
	}
	go func(t *SimpleThread[I, O]) {
		for {
			t.proclock.Lock()
			t.proclock.Unlock()
			select {
			case i := <-inchan:
				outchan <- t.fn(i)
			case <-diechan:
				close(outchan)
				close(inchan)
				close(diechan)
				t.flags ^= running
				return
			}
		}
	}(&thr)
	return &thr
}

func NewSimple[I, O any](
	fn func(I) O,
	inlen int,
	outlen int,
) *SimpleThread[I, O] {
	inchan := make(chan I, inlen)
	outchan := make(chan O, outlen)
	thr := ContainSimple(fn, inchan, outchan)
	thr.flags |= worker
	return thr
}

// impl Pipe
func (t *SimpleThread[I, O]) Send(i I) {
	t.iolock.Lock()
	t.iolock.Unlock()
	t.inchan <- i
}
func (t *SimpleThread[I, O]) Await() O {
	t.iolock.Lock()
	t.iolock.Unlock()
	return <-t.outchan
}
func (t *SimpleThread[I, O]) Die() {
	t.diechan <- struct{}{}
}
func (t *SimpleThread[I, O]) InQueue() int {
	return len(t.inchan)
}
func (t *SimpleThread[I, O]) InMaxQueue() int {
	return cap(t.inchan)
}
func (t *SimpleThread[I, O]) OutQueue() int {
	return len(t.outchan)
}
func (t *SimpleThread[I, O]) OutMaxQueue() int {
	return cap(t.outchan)
}
func (t *SimpleThread[I, O]) LockIO() (chan I, chan O) {
	t.iolock.Lock()
	t.state = locked
	return t.inchan, t.outchan
}
func (t *SimpleThread[I, O]) Own() (chan I, chan O) {
	t.state = threadOwned
	return t.inchan, t.outchan
}
func (t *SimpleThread[I, O]) LockProc() {
	t.proclock.Lock()
}
func (t *SimpleThread[I, O]) UnlockProc() {
	t.proclock.Unlock()
}
func (t *SimpleThread[I, O]) UnlockIO(chan I, chan O) {
	t.state = userOwned
	t.iolock.Unlock()
}
