package main

//$ go mod init main
//$ go mod tidy

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
	WaitTrigger      = 120
	TimeCommand      = "echo $((`ioreg -c IOHIDSystem | sed -e '/HIDIdleTime/ !{ d' -e 't' -e '}' -e 's/.* = //g' -e 'q'` / 1000000000))"
	LockScreenScript = `import sys,Quartz
try:
	print(Quartz.CGSessionCopyCurrentDictionary())
except:
	print(".")`
)

type Waiter interface {
	Init()
	IsScreenLocked() bool
	IncrementDelay()
	ResetDelay()
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
	s.ResetDelay()
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
		if idle := stringToInt64(sb.String()); s.ShouldNotify(idle) {
			c <- 1
			s.ResetDelay()
		}
		time.Sleep(time.Duration(s.delay) * time.Second)
	}
}

func (s *Sleeper) ShouldNotify(idle int64) bool {
	if s.IsScreenLocked() {
		s.IncrementDelay()
		return false
	}
	if idle < WaitTrigger {
		s.IncrementDelay()
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
	return true
}

func (s *Sleeper) IncrementDelay() {
	if s.delay < WaitTrigger/2 {
		s.delay += 1
	}
}

func (s *Sleeper) ResetDelay() {
	s.delay = 2
}

func (s *Sleeper) Move(c <-chan byte) {
	for range c {
		x, y := robotgo.GetMousePos()
		robotgo.MoveSmooth(x-2+rand.Intn(4), y-2+rand.Intn(4))
	}
}

func stringToInt64(str string) int64 {
	idle, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
	if err != nil {
		log.Fatalln("Conversion error: ", err)
	}
	return idle
}
