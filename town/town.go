package town

import (
	"fmt"
)

type Thread[I, O any] struct {
	output  chan O
	input   chan I
	die     chan struct{}
	running bool
}

func (t *Thread[I, O]) ExecSimple(f func(exec I) O ) error {
	if t.running {
		return fmt.Errorf("already running")
	} else {
		t.running = true
		nf := func(exec <-chan I, yield chan<- O, die chan struct{}) {
			for {
				select {
				case i := <-exec:
					yield<-f(i)
				case <-die:
					return
				}
			}
		}
		go nf(t.input, t.output, t.die)
		return nil
	}
}

func (t *Thread[I, O]) Exec(f func(exec <-chan I, yield chan<- O, die chan struct{})) error {
	if t.running {
		return fmt.Errorf("already running")
	} else {
		t.running = true
		go f(t.input, t.output, t.die)
		return nil
	}
}

func (t *Thread[I, O]) Send(i I) {
	t.input <- i
}
func (t *Thread[I, O]) Await() O {
	return <-t.output
}
func (t *Thread[I, O]) AwaitAll() chan O {
	c := make(chan O, len(t.input)+len(t.output))
	for i := 0; i < cap(c); i++ {
		c <- <-t.output
	}
	close(c)
	return c
}
func (t *Thread[I, O]) Run(i I) O {
	t.Send(i)
	return t.Await()
}
func (t Thread[I, O]) Die() chan O {
	t.running = false
	t.die <- struct{}{}
	close(t.die)
	close(t.output)
	close(t.input)
	return t.output
}
func (t Thread[I, O]) DieAwait() chan O {
	t.running = false
	c := t.AwaitAll()
	t.die <- struct{}{}
	close(t.die)
	close(t.output)
	close(t.input)
	return c
}
func (t *Thread[I, O]) Act() (snd chan<- I, rcv <-chan O) {
	return t.input, t.output
}

func New[I, O any](inlen, outlen int) *Thread[I, O] {
	output := make(chan O, outlen)
	input := make(chan I, inlen)
	die := make(chan struct{})
	running := false
	return &Thread[I, O]{
		output,
		input,
		die,
		running,
	}
}
