package main

import (
	"fmt"
	gp "project/pool"
	"time"
)


func main() {
	t := gp.NewTask(
		func() error {
			fmt.Println(time.Now())
			return nil
		})

	p := gp.NewGoroutinePool(4)
	go func() {
		for {
			p.AddSingleTask(t)
			
		}
	}()
	p.Run()
}