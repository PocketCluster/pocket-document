package main

import (
    "fmt"
    "math/rand"
    "time"
    "sync"

    log "github.com/Sirupsen/logrus"
)

// non-deterministic choice
// https://stackoverflow.com/questions/31539804/multiple-receivers-on-a-single-channel-who-gets-the-data
// https://stackoverflow.com/questions/18660533/why-using-unbuffered-channel-in-the-the-same-goroutine-gives-a-deadlock


func reader1(task <- chan string, stat chan <- int, exit <- chan struct{}, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case data := <- task:
            log.Printf("reader 1 : %v", data)
            time.Sleep(time.Millisecond * time.Duration(100 * rand.Intn(10)))
            stat <- 0
        case <- exit:
            return
        }
    }
}

func reader2(task <- chan string, stat chan <- int, exit <- chan struct{}, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case data := <- task:
            log.Printf("reader 2 : %v", data)
            time.Sleep(time.Millisecond * time.Duration(100 * rand.Intn(10)))
            stat <- 1
        case <- exit:
            return
        }
    }
}

func reader3(task <- chan string, stat chan <- int, exit <- chan struct{}, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case data := <- task:
            log.Printf("reader 3 : %v", data)
            time.Sleep(time.Millisecond * time.Duration(100 * rand.Intn(10)))
            stat <- 2
        case <- exit:
            return
        }
    }
}

func reader4(task <- chan string, stat chan <- int, exit <- chan struct{}, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case data := <- task:
            log.Printf("reader 4 : %v", data)
            time.Sleep(time.Millisecond * time.Duration(100 * rand.Intn(10)))
            stat <- 3
        case <- exit:
            return
        }
    }
}

func main() {

    var (
        waiter = sync.WaitGroup{}
        task = make(chan string)
        stat = make(chan int)
        exit = make(chan struct{})
        record = []uint{0, 0, 0, 0}
    )

    go reader1(task, stat, exit, &waiter)
    go reader2(task, stat, exit, &waiter)
    go reader3(task, stat, exit, &waiter)
    go reader4(task, stat, exit, &waiter)


    go func(rec []uint) {
        for {
            select {
                case s := <- stat:
                    rec[s]++
                case <- exit:
                    return
            }
        }
    }(record)

    for i := 0; i < 1000; i++ {
        task <- fmt.Sprintf("Did you copy? %d", i)
    }

    close(exit)
    waiter.Wait()
    close(task)
    close(stat)
    for j := range record {
        log.Printf("Channel count %d : %d", j + 1, record[j])
    }
}
