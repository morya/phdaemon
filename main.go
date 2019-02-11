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

	"github.com/morya/utils/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var (
	flagSleep = flag.Int("sleep", 10, "sleep second")
    flagAlertUrl = flag.String("alertUrl", "https://sc.ftqq.com/SCU20246T6f4f873407163ae0f49ca8ca8788c3385b40c18a19d6f.send", "")
)

type Sys struct {
	wg   *sync.WaitGroup
	stop chan int
}

func NewSys() *Sys {
	return &Sys{
		wg:   new(sync.WaitGroup),
		stop: make(chan int),
	}
}

func (*Sys) checkMem() {
	v, _ := mem.VirtualMemory()
	s := fmt.Sprintf("mem used percent:%.2f%%", v.UsedPercent)
	log.Info(s)
	if v.UsedPercent > 25 {
		fangtangNotify("mem alert", s)
	}
}

func (*Sys) checkCpu() {
	v, _ := cpu.Percent(time.Millisecond*300, false)
	f := v[0]
	s := fmt.Sprintf("cpu used percent:%.2f%%", f)
	log.Info(s)
	if f > 20 {
		fangtangNotify("cpu alert", s)
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
		log.Info("will start oraysl")
		var args = strings.Split("-a 127.0.0.1 -p 16062 -s phsle01.oray.net:6061 -l /var/log/phddns -L oraysl -d", " ")
		cmd := exec.Command("/usr/orayapp/oraysl", args...)
		cmd.Start()
	} else {
		log.Infof("oraysl pid = %v", ps.Pid)
	}
}

func (s *Sys) checkOraynewph(processes []*process.Process) {
	var ps = s.FindProcess(processes, "oraynewph")
	if ps == nil {
		log.Info("will start oraynewph")
		var args = strings.Split("-s 0.0.0.0  -c /var/log/phddns/core.log -p /var/log/phddns -l oraynewph", " ")
		cmd := exec.Command("/usr/orayapp/oraynewph", args...)
		cmd.Start()
	} else {
		log.Infof("oraynewph pid = %v", ps.Pid)
	}
}

func (s *Sys) Stop() {
	timeout := time.NewTicker(time.Millisecond * 50)
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

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Info("started")

	s := NewSys()
	go s.Run()
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	s.Stop()
}
