package main

import (
    "github.com/go-vgo/robotgo"
    "log"
    "math/rand"
    "os/exec"
    "strconv"
    "strings"
    "sync"
    "time"
)

const timeCommand = "echo $((`ioreg -c IOHIDSystem | sed -e '/HIDIdleTime/ !{ d' -e 't' -e '}' -e 's/.* = //g' -e 'q'` / 1000000000))"
const idleTrigger = 120

func main() {
    wg := sync.WaitGroup{}
    wg.Add(1)
    pipe := make(chan byte, 4)
    go notifyWhenIdle(pipe)
    go moveMouse(pipe)
    wg.Wait()
}

func moveMouse(c <-chan byte) {
    for range c {
        x, y := robotgo.GetMousePos()
        robotgo.MoveMouseSmooth(x - 2 + rand.Intn(4), y - 2 + rand.Intn(4))
    }
}

func notifyWhenIdle(c chan<- byte) {
    sb := new(strings.Builder)
    for {
        sb.Reset()
        cmd := exec.Command("/bin/sh", "-c", timeCommand)
        cmd.Stdout = sb
        if err := cmd.Run(); err != nil {
            log.Fatalln("Command Run() error: ", err)
        }
        if idle := convertToInt(sb); idle > idleTrigger {
            c <- 1
        }
        time.Sleep(2 * time.Second)
    }
}

func convertToInt(sb *strings.Builder) int64 {
    idle, err := strconv.ParseInt(strings.TrimSpace(sb.String()), 10, 64)
    if err != nil {
        log.Fatalln("Conversion error: ", err)
    }
    return idle
}
