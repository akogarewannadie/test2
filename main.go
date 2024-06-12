package main

import (
	"fmt"
	"sync"
	"time"
)

type Ttype struct {
	id         int
	cT         string
	fT         string
	taskRESULT string
}

func main() {

	taskChan := make(chan Ttype, 10)

	doneTasks := make(chan Ttype, 10)
	undoneTasks := make(chan error, 10)

	var wg sync.WaitGroup
	var mu sync.Mutex

	taskCreator := func() {
		defer close(taskChan)
		end := time.Now().Add(10 * time.Second)
		for time.Now().Before(end) {
			ft := time.Now().Format(time.RFC3339)
			if time.Now().Nanosecond()%2 > 0 {
				ft = "Some error occurred"
			}
			taskChan <- Ttype{cT: ft, id: int(time.Now().UnixNano())}
			time.Sleep(100 * time.Millisecond)
		}
	}

	taskWorker := func(t Ttype) Ttype {
		tt, err := time.Parse(time.RFC3339, t.cT)
		if err != nil {
			t.taskRESULT = "parsing error"
		} else if tt.After(time.Now().Add(-20 * time.Second)) {
			t.taskRESULT = "task has been successful"
		} else {
			t.taskRESULT = "something went wrong"
		}
		t.fT = time.Now().Format(time.RFC3339Nano)
		time.Sleep(150 * time.Millisecond)
		return t
	}

	taskSorter := func(t Ttype) {
		if t.taskRESULT == "task has been successful" {
			doneTasks <- t
		} else {
			undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, t.taskRESULT)
		}
	}

	go taskCreator()

	go func() {
		for t := range taskChan {
			wg.Add(1)
			go func(task Ttype) {
				defer wg.Done()
				processedTask := taskWorker(task)
				taskSorter(processedTask)
			}(t)
		}
		wg.Wait()
		close(doneTasks)
		close(undoneTasks)
	}()

	go func() {
		for {
			time.Sleep(3 * time.Second)
			mu.Lock()
			fmt.Println("Errors:")
			for len(undoneTasks) > 0 {
				fmt.Println(<-undoneTasks)
			}
			fmt.Println("Done tasks:")
			for len(doneTasks) > 0 {
				fmt.Println(<-doneTasks)
			}
			mu.Unlock()
		}
	}()

	time.Sleep(13 * time.Second)
}
