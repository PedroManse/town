package main

import (
	"fmt"
	"town/town"
)

func InitDoubler() *town.Thread[int, int] {
	t := town.New[int, int](140, 140)
	double := func(input <-chan int, yield chan<- int, die chan struct{}) {
		for {
			select {
			case i := <-input:
				yield <- 2 * i
			case <-die:
				for i := range input {
					yield <- 2*i
				}
				return
			}
		}
	}
	t.Exec(double)
	return t
}
func InitSimpleDoubler() *town.Thread[int, int] {
	t := town.New[int, int](140, 140)
	double := func(input int) int {
		return input*2
	}
	t.ExecSimple(double)
	return t
}

func main() {
	dblr := InitDoubler()
	snd, _ := dblr.Act()
	snd <- 100
	snd <- 100
	snd <- 100
	snd <- 100
	snd <- 100
	for out := range dblr.Die() {
		fmt.Println(out)
	}
}
