package main

import (
	"fmt"
	gp "project/pool"
	"time"
	"os/signal"
	"os"
)


func main() {
	t := gp.NewTask(
		func() error {
			// fmt.Println(time.Now())
			// 耗时代表
			time.Sleep(5000)
			return nil
		})

	p := gp.NewGoroutinePool(4)
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, os.Kill)
	exitFlag := make(chan bool, 1)
	go func() {
		for {
			select{
			case <-exitFlag:
				break
			default:
				p.AddSingleTask(t)
			}
		}
		fmt.Println("sendStop")
	}()
	p.Run()
	<- exitChan
	fmt.Println("os want to stop")
	exitFlag <- true
	p.Stop()
}