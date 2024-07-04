package main

import (
	"fmt"
	//"time"
	"town/town2"
)

func main() {
	dblr := town2.NewSimple(
		func(i int) int { return i * 2 },
		5, 5,
	)
	quadr := town2.NewSimple(
		func(i int) int { return i * 4 },
		5, 5,
	)
	oct := town2.Join(dblr, quadr)

	oct.Send(1)
	oct.Send(1)
	oct.Await()
	y := town2.SinkThread(oct)

	fmt.Printf("%+v\n", y)
}
