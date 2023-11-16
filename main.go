package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
)

const (
	DINGDING_ROBOT_URL = ""
)

var (
	flagSleep    = flag.Int("sleep", 10, "sleep second")
	flagAlertUrl = flag.String("alertUrl", DINGDING_ROBOT_URL, "")
	flagLogLevel = flag.String("logLevel", "info", "log level")
)

type Sys struct {
	cpuAlertCount int
	memAlertCount int
	wg            *sync.WaitGroup
	stop          chan int
}

func NewSys() *Sys {
	return &Sys{
		wg:   new(sync.WaitGroup),
		stop: make(chan int),
	}
}

func (ss *Sys) checkMem() {
	v, _ := mem.VirtualMemory()
	s := fmt.Sprintf("mem used percent:%.2f%%", v.UsedPercent)
	logrus.Info(s)
	if v.UsedPercent > 75 {
		ss.memAlertCount++
		if ss.memAlertCount > 3 {
			sendAlert("mem alert", s)
			ss.memAlertCount = 0
		}
	} else {
		ss.memAlertCount = 0
	}
}

func (ss *Sys) checkCpu() {
	v, _ := cpu.Percent(time.Millisecond*300, false)
	f := v[0]
	s := fmt.Sprintf("cpu used percent:%.2f%%", f)
	logrus.Info(s)
	if f > 20 {
		ss.cpuAlertCount++
		if ss.cpuAlertCount > 3 {
			sendAlert("cpu alert", s)
			ss.cpuAlertCount = 0
		}
	} else {
		ss.cpuAlertCount = 0
	}
}

func (s *Sys) FindProcess(processes []*process.Process, desc string) *process.Process {
	var ps *process.Process
	for _, p := range processes {
		exe, _ := p.Exe()
		if -1 != strings.Index(exe, desc) {
			ps = p
			break
		}
	}
	return ps
}

func (s *Sys) checkOraysl(processes []*process.Process) {
	var ps = s.FindProcess(processes, "oraysl")
	if ps == nil {
		logrus.Info("will start oraysl")
		var args = strings.Split("-a 127.0.0.1 -p 16062 -s phsle01.oray.net:6061 -l /var/log/phddns -L oraysl -d", " ")
		cmd := exec.Command("/usr/orayapp/oraysl", args...)
		cmd.Start()
	} else {
		logrus.Infof("oraysl pid = %v", ps.Pid)
	}
}

func (s *Sys) checkOraynewph(processes []*process.Process) {
	var ps = s.FindProcess(processes, "oraynewph")
	if ps == nil {
		logrus.Info("will start oraynewph")
		var args = strings.Split("-s 0.0.0.0  -c /var/log/phddns/core.log -p /var/log/phddns -l oraynewph", " ")
		cmd := exec.Command("/usr/orayapp/oraynewph", args...)
		cmd.Start()
	} else {
		logrus.Infof("oraynewph pid = %v", ps.Pid)
	}
}

func (s *Sys) Stop() {
	timeout := time.NewTicker(time.Millisecond * 50)
	defer timeout.Stop()
	select {
	case s.stop <- 1:
	case <-timeout.C:
	}

	s.wg.Wait()
}

func (s *Sys) Run() {
	s.wg.Add(1)
	defer s.wg.Done()
	ticker := time.NewTicker(time.Duration(*flagSleep) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.checkCpu()
			s.checkMem()

			processes, err := process.Processes()
			if err != nil {
				continue
			}
			s.checkOraysl(processes)
			s.checkOraynewph(processes)

		case <-s.stop:
			return
		}
	}
}

func main() {
	flag.Parse()

	l, _ := logrus.ParseLevel(*flagLogLevel)
	logrus.SetLevel(l)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, DisableQuote: true})

	logrus.Info("started")

	s := NewSys()
	go s.Run()
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	s.Stop()
}
