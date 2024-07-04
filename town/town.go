package town

import (
	"fmt"
)

type Thread[I, O any] struct {
	output  chan O
	input   chan I
	die     chan struct{}
	running bool
	ghost   bool
}

func (t *Thread[I, O]) Send(i I) {
	fmt.Println("should be ghost", t)
	fmt.Println(len(t.input), len(t.output))
	t.input <- i
	fmt.Println(len(t.input), len(t.output))
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

func (t *Thread[I, O]) Die() chan O {
	t.running = false
	t.die <- struct{}{}
	close(t.die)
	close(t.output)
	close(t.input)
	return t.output
}

func (t *Thread[I, O]) DieAwait() chan O {
	c := t.AwaitAll()
	t.running = false
	t.die <- struct{}{}
	close(t.die)
	close(t.output)
	close(t.input)
	return c
}

func (t *Thread[I, O]) Avail() chan O {
	c := make(chan O, len(t.output))
	for i := 0; i < cap(c); i++ {
		c <- <-t.output
	}
	close(c)
	return c
}

func (t *Thread[I, O]) Unpack() (snd chan<- I, rcv <-chan O) {
	return t.input, t.output
}

func WrapSimple[I, O any](
	f func(exec I) O,
) (
	func(exec <-chan I, yield chan<- O, die <-chan struct{}),
) {
	return func(input <-chan I, yield chan<- O, die <-chan struct{}) {
		for {
			select {
			case i := <-input:
				yield <- f(i)
			case <-die:
				return
			}
		}
	}
}

func WrapSink[I any](
	f func(exec I),
) (
	func(exec <-chan I, yield chan<- struct{}, die <-chan struct{}),
) {
	return func(input <-chan I, yield chan<- struct{}, die <-chan struct{}) {
		for {
			select {
			case i := <-input:
				f(i)
				yield <- struct{}{}
			case <-die:
				return
			}
		}
	}
}

// bad
func Pipe[I, IO, O any](t1 *Thread[I, IO], t2 *Thread[IO, O]) *Thread[I, O] {
	c := ContainChans(t1.input, t2.output, make(chan struct{}))
	c.ghost = true // doesn't actually do work
	fmt.Println("is ghost", c)
	c.Exec(
		func(exec <-chan I, yield chan<- O, die <-chan struct{}) {
			for { select {
			case i:=<-exec:
				fmt.Println(i)
				yield<-t2.Run(t1.Run(i))
			case <-die:
				c1 := t1.AwaitAll()
				t1.running = false
				t1.die <- struct{}{}
				for c := range c1 {
					t2.Send(c)
				}
				_=t2.AwaitAll()
				t2.running = false
				t2.die <- struct{}{}
				close(t1.input)
				close(t2.output)
			} }
		},
	)
	return c
}

// good
func Append[I, IO, O any](
	t1 *Thread[I, IO], outlen int,
	f func(exec <-chan IO, yield chan<- O, die <-chan struct{}),
) *Thread[I, O]{
	output := make(chan O, outlen)
	die := make(chan struct{})
	t2 := ContainChans(t1.output, output, die)
	// it's ok do put die chan here
	// since tout is a ghost Thread, nothing in tout will read it
	// but t2 will
	tout := ContainChans(t1.input, t2.output, die)
	tout.running = true
	tout.ghost = true
	t2.Exec(f)
	return tout
}

// good2
func Join[I, IO, O any] (
	f1 func(exec I) IO,
	f2 func(exec IO) O,
) func(exec I) O {
	return func(exec I) O {
		return f2(f1(exec))
	}
}

func New[I, O any](
	inlen, outlen int,
	f func(exec <-chan I, yield chan<- O, die <-chan struct{}),
) (t *Thread[I, O]) {
	t = CreateThread[I, O](inlen, outlen)
	t.Exec(f)
	return
}

func NewSimple[I, O any](
	inlen, outlen int,
	f func(exec I) O,
) (t *Thread[I, O]) {
	t = CreateThread[I, O](inlen, outlen)
	t.ExecSimple(f)
	return
}

func ContainChans[I, O any](
	input chan I,
	output chan O,
	die chan struct{},
) *Thread[I, O] {
	running := false
	return &Thread[I, O]{
		output: output,
		input: input,
		die: die,
		running: running,
	}
}

func CreateThread[I, O any](inlen, outlen int) *Thread[I, O] {
	output := make(chan O, outlen)
	input := make(chan I, inlen)
	die := make(chan struct{})
	running := false
	return &Thread[I, O]{
		output: output,
		input: input,
		die: die,
		running: running,
	}
}

func (t *Thread[I, O]) ExecSimple(f func(exec I) O) error {
	if t.running {
		return fmt.Errorf("already running")
	}
	t.running = true
	go func(exec <-chan I, yield chan<- O, die <-chan struct{}) {
		for {
			select {
			case i := <-t.input:
				t.output <- f(i)
			case <-t.die:
				return
			}
		}
	}(t.input, t.output, t.die)
	return nil
}

func (t *Thread[I, O]) Exec(f func(exec <-chan I, yield chan<- O, die <-chan struct{})) error {
	if t.running {
		return fmt.Errorf("already running")
	}
	t.running = true
	go f(t.input, t.output, t.die)
	return nil
}

