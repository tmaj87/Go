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

const (
	WaitTrigger = 120
	TimeCommand = "echo $((`ioreg -c IOHIDSystem | sed -e '/HIDIdleTime/ !{ d' -e 't' -e '}' -e 's/.* = //g' -e 'q'` / 1000000000))"
	LockScreenScript = `import sys,Quartz
try:
	print(Quartz.CGSessionCopyCurrentDictionary())
except:
	print(".")`
)

type Waiter interface {
	Init()
	IsScreenLocked() bool
	Increment()
	Reset()
	Move(c <-chan byte)
	Notify(c chan<- byte)
	ShouldNotify(idle int64) bool
}

type Sleeper struct {
	delay int64
}

// var _ Waiter = Sleeper{}
var _ Waiter = (*Sleeper)(nil)

func main() {
	s := &Sleeper{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	s.Init()
	wg.Wait()
}

func (s *Sleeper) Init() {
	s.Reset()
	c := make(chan byte, 4)
	go s.Notify(c)
	go s.Move(c)
}

func (s *Sleeper) Notify(c chan<- byte) {
	sb := new(strings.Builder)
	for {
		sb.Reset()
		cmd := exec.Command("/bin/sh", "-c", TimeCommand)
		cmd.Stdout = sb
		if err := cmd.Run(); err != nil {
			log.Fatalln("Command Run() error: ", err)
		}
		if idle := convertToInt(sb); s.ShouldNotify(idle) {
			c <- 1
			s.Reset()
		}
		time.Sleep(time.Duration(s.delay) * time.Second)
	}
}

func (s *Sleeper) ShouldNotify(idle int64) bool {
	if s.IsScreenLocked() {
		return false
	}
	if idle < WaitTrigger {
		return false
	}
	return true
}

func (s *Sleeper) IsScreenLocked() bool {
	sb := new(strings.Builder)
	cmd := exec.Command("python3", "-c", LockScreenScript)
	cmd.Stdout = sb
	if err := cmd.Run(); err != nil {
		log.Fatalln("Command Run() error: ", err)
	}
	if !strings.Contains(sb.String(), "CGSSessionScreenIsLocked = 1") {
		return false
	}
	s.Increment()
	return true
}

func (s *Sleeper) Increment() {
	if s.delay < WaitTrigger {
		s.delay += 1
	}
}

func (s *Sleeper) Reset() {
	s.delay = 2
}

func (s *Sleeper) Move(c <-chan byte) {
	for range c {
		x, y := robotgo.GetMousePos()
		robotgo.MoveSmooth(x-2+rand.Intn(4), y-2+rand.Intn(4))
	}
}

func convertToInt(sb *strings.Builder) int64 {
	idle, err := strconv.ParseInt(strings.TrimSpace(sb.String()), 10, 64)
	if err != nil {
		log.Fatalln("Conversion error: ", err)
	}
	return idle
}
