package town2

type Pipe[I, O any] interface {
	Own() (chan I, chan O)
	LockIO() (chan I, chan O)
	UnlockIO(chan I, chan O)
	LockProc()
	UnlockProc()
	Send(I)
	Await() O
	Die()
	InQueue() int
	InMaxQueue() int
	OutQueue() int
	OutMaxQueue() int
}

func Is(flags, is ThreadFlag) bool {
	return flags&is != 0
}

type ThreadFlag uint64

const (
	_ ThreadFlag = 1 << iota
	ghost
	simple  // !simple = long running
	running // !running = dead
	worker  // !worker = piper
)

type ThreadState uint64

const (
	_ ThreadState = iota
	userOwned
	threadOwned
	locked
)

func SinkThread[I, O any](p Pipe[I, O]) []O {
	in, out := p.LockIO()
	qout := len(out)
	os := make([]O, qout)
	for i := range qout {
		os[i] = <-out
	}
	p.UnlockIO(in, out)
	return os
}

func Join[I, IO, O any](
	p1 Pipe[I, IO], p2 Pipe[IO, O],
) Pipe[I, O] {
	userIn, ownedOut := p1.Own()
	ownedIn, userOut := p2.Own()
	piper := ContainSimple[IO, IO](
		func(i IO) IO {
			return i
		},
		ownedIn, ownedOut,
	)
	return NewGhost[I, IO, O](
		p1, piper, p2,
		userIn, userOut,
	)
}
