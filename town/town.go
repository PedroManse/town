package town

type Thread[I, O any] struct {
	yield chan O;
	exec chan I;
	die chan struct{};
}

func New[I, O]() *Thread[I, O] {
	Thread{
		yield: make(chan O),
		exec: make(chan I),
		die: make(chan struct{}),
	}
}
