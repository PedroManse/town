package main

import (
	"fmt"
	"reflect"
	"runtime"
	"time"
	"town/town"
)

func InitDoubler() *town.Thread[int, int] {
	dblr := func(input <-chan int, yield chan<- int, die <-chan struct{}) {
		for {
			select {
			case i := <-input:
				yield <- 2 * i
			case <-die:
				for i := range input {
					yield <- 2 * i
				}
				return
			}
		}
	}
	t := town.New[int, int](140, 140, dblr)
	return t
}

func InitSimpleDoubler() *town.Thread[int, int] {
	return town.NewSimple[int, int](140, 140,
		func(input int) int {
			return input * 2
		},
	)
}

// +1c
func InitQuadGood2() *town.Thread[int, struct{}] {
	dblr := func(i int) int {
		return i * 2
	}
	prt := func(i int) struct{} {
		fmt.Println(i)
		return struct{}{}
	}

	x := town.Join(dblr, dblr)
	y := town.Join(x, prt)
	t := town.New[int, struct{}](100, 100, town.WrapSimple(y)) // +1c
	return t
}

// +3c
func InitQuadGood() *town.Thread[int, struct{}] {
	dblr := town.WrapSimple(
		func(i int) int {
			return i * 2
		},
	)
	prt := town.WrapSink(
		func(i int) {
			fmt.Println(i)
		},
	)

	t1 := town.New[int, int](100, 100, dblr)            // +1c
	t2 := town.Append[int, int, int](t1, 100, dblr)     // +1c
	to := town.Append[int, int, struct{}](t2, 100, prt) // +1c
	return to
}

// +5c
func InitQuadBad() *town.Thread[int, struct{}] {
	dblr := town.WrapSimple(
		func(i int) int {
			return i * 2
		},
	)
	prt := town.WrapSink(
		func(i int) {
			fmt.Println(i)
		},
	)

	//t2 := town.New[int, int](100, 100, dblr)   // +1c
	t1 := town.New[int, int](100, 100, dblr)     // +1c
	t3 := town.New[int, struct{}](100, 100, prt) // +1c
	t1_t3 := town.Pipe(t1, t3)                   // +1c
	//t1t2_t3 := town.Pipe(t1_t2, t3)            // +1c
	return t1_t3
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func testFN(fn func() *town.Thread[int, struct{}]) {
	fname := GetFunctionName(fn)
	countStart := runtime.NumGoroutine()
	TO := fn()
	countEnd := runtime.NumGoroutine()
	fmt.Println("is user's", TO)
	fmt.Printf("The Owned Thread %q spawned %d goroutines\n", fname, countEnd-countStart)
	TO.Run(3)
	TO.DieAwait()
	fmt.Printf("done testing\n")
}

func main() {
	testFN(InitQuadBad)
	testFN(InitQuadGood)
	testFN(InitQuadGood2)
	time.Sleep(time.Second)
}
