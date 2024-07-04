package town2

import (
)

// TODO make actual ghost thread
// that does no work
// make Piper thread that works like current Ghost

type GhostThread[I, IO, O any] struct {
	state   ThreadState
	flags   ThreadFlag
	inPipe  Pipe[I, IO]
	piper   *SimpleThread[IO, IO]
	outPipe Pipe[IO, O]
	uin     chan I
	uout    chan O
	diechan chan struct{}
}

func NewGhost[I, IO, O any](
	inpipe Pipe[I, IO], // working thread
	piper *SimpleThread[IO, IO], // piping thread
	outpipe Pipe[IO, O], // working thread
	uin chan I,
	uout chan O,
) *GhostThread[I, IO, O] {
	diechan := make(chan struct{})
	piper.Own()

	thr := GhostThread[I, IO, O]{
		state:   userOwned,
		flags:   ghost + running,
		inPipe:  inpipe,
		piper:   piper,
		outPipe: outpipe,
		diechan: diechan,
		uin:     uin,
		uout:    uout,
	}
	return &thr
}

//var LINMAP = syncMap[uintptr, any]{
//	mp: map[uintptr]any{},
//	mx: sync.Mutex{},
//}

type ILIN[I, IO, O any] struct {
	inI  chan I
	inO  chan IO
	outI chan IO
	outO chan O
}

// lock both IO ends
func (g *GhostThread[I, IO, O]) IOLK() ILIN[I, IO, O] {
	inI, inO := g.inPipe.LockIO()
	outI, outO := g.outPipe.LockIO()
	return ILIN[I, IO, O]{
		inI, inO,
		outI, outO,
	}
}
func (g *GhostThread[I, IO, O]) IOUL(LN *ILIN[I, IO, O]) {
	//g.inPipe.UnlockIO(LN.inI, LN.inO)
	//g.outPipe.UnlockIO(LN.outI, LN.outO)
	g.inPipe.UnlockIO(nil,nil)
	g.outPipe.UnlockIO(nil,nil)
}

func (g *GhostThread[I, IO, O]) Send(i I) {
	g.inPipe.Send(i)
}
func (g *GhostThread[I, IO, O]) Await() O {
	return g.outPipe.Await()
}
func (g *GhostThread[I, IO, O]) LockIO() (in chan I, out chan O) {
	LN := g.IOLK()
	return LN.inI, LN.outO
}

func (g *GhostThread[I, IO, O]) UnlockIO(in chan I, out chan O) {
	g.IOUL(nil)
}

func (g *GhostThread[I, IO, O]) Die() {
	g.inPipe.Die()
	g.piper.Die()
	g.outPipe.Die()
}

func (g *GhostThread[I, IO, O]) InMaxQueue() int {
	return g.inPipe.InMaxQueue()
}
func (g *GhostThread[I, IO, O]) InQueue() int {
	return g.inPipe.InQueue()
}
func (g *GhostThread[I, IO, O]) OutQueue() int {
	return g.outPipe.OutQueue()
}
func (g *GhostThread[I, IO, O]) OutMaxQueue() int {
	return g.outPipe.OutMaxQueue()
}
func (g *GhostThread[I, IO, O]) LockProc() {
	g.inPipe.LockProc()
	g.outPipe.LockProc()
}
func (g *GhostThread[I, IO, O]) UnlockProc() {
	g.inPipe.UnlockProc()
	g.outPipe.UnlockProc()
}
func (t *GhostThread[I, IO, O]) Own() (chan I, chan O) {
	t.state = threadOwned
	return t.uin, t.uout
}
