package main

import (
	"context"
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

var sleepDelay = 2

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
		robotgo.MoveSmooth(x-2+rand.Intn(4), y-2+rand.Intn(4))
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
		if idle := convertToInt(sb); shouldNotify(idle) {
			c <- 1
			resetSleep()
		}
		time.Sleep(time.Duration(sleepDelay) * time.Second)
	}
}

func increaseSleep() {
	if sleepDelay < idleTrigger {
		sleepDelay += 1
	}
}

func resetSleep() {
	sleepDelay = 2
}

func shouldNotify(idle int64) bool {
	if screenIsLocked() {
		return false
	}
	if idle < idleTrigger {
		return false
	}
	hour := time.Now().Hour()
	if hour < 8 || hour > 20 {
		return false
	}
	return true
}

func screenIsLocked() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	cmd := exec.CommandContext(
		ctx,
		"python3",
		"-c",
		"import sys,Quartz; d=Quartz.CGSessionCopyCurrentDictionary(); print(d)",
	)

	var b []byte
	var err error
	if b, err = cmd.CombinedOutput(); err != nil {
		cancel()
		log.Fatalln("CGSessionCopyCurrentDictionary error: ", err)
		return false
	}
	cancel()

	if !strings.Contains(string(b), "CGSSessionScreenIsLocked = 1") {
		return false
	}
	increaseSleep()
	return true
}

func convertToInt(sb *strings.Builder) int64 {
	idle, err := strconv.ParseInt(strings.TrimSpace(sb.String()), 10, 64)
	if err != nil {
		log.Fatalln("Conversion error: ", err)
	}
	return idle
}
