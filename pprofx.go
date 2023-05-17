package pprofx

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"syscall"
	"time"
)

type (
	cpuTask struct {
		state   cpuState
		profile *os.File
		heap    *os.File
	}
	cpuState string
)

const (
	CPUStateWAITING  = cpuState("waiting")
	CPUStateACTIVE   = cpuState("active")
	CPUStateFINISHED = cpuState("finished")
	CPUStateIDLE     = cpuState("idle")
)

func (cpu *cpuTask) log(info ...interface{}) {
	fmt.Println(append([]interface{}{time.Now().Format("2006-01-02 15:04:05"), "[pprofx]"}, info...)...)
}

func (cpu *cpuTask) CheckState(next cpuState) (err error) {
	flag := false
	switch cpu.state {
	case CPUStateWAITING:
		flag = next == CPUStateACTIVE
	case CPUStateACTIVE:
		flag = next == CPUStateFINISHED
	case CPUStateFINISHED:
		flag = next == CPUStateIDLE
	case CPUStateIDLE:
		flag = next == CPUStateWAITING
	}
	if !flag {
		err = fmt.Errorf("current state %s canot to be %s", cpu.state, next)
	}
	return
}

// CreateFile 打开要写入性能分析数据的文件
func (cpu *cpuTask) CreateFile(name string) error {
	var (
		err  error
		path string
	)
	if path, err = os.Getwd(); err != nil {
		return err
	}

	if err = cpu.CheckState(CPUStateWAITING); err != nil {
		return err
	}
	sts := time.Now().Format("20060102150405")
	cpu.profile, err = os.Create(filepath.Join(path, fmt.Sprintf("%s_%s.profile", name, sts)))
	if err != nil {
		return err
	}
	cpu.heap, err = os.Create(filepath.Join(path, fmt.Sprintf("%s_%s.heap", name, sts)))
	if err != nil {
		return err
	}
	cpu.state = CPUStateWAITING
	cpu.log("profile file created", cpu.profile.Name())
	cpu.log("heap file created", cpu.heap.Name())
	return nil
}

// StartRecording 开始 CPU 分析
func (cpu *cpuTask) StartRecording() error {
	if err := cpu.CheckState(CPUStateACTIVE); err != nil {
		return err
	}
	if err := pprof.StartCPUProfile(cpu.profile); err != nil {
		return err
	}
	cpu.state = CPUStateACTIVE
	cpu.log("start recording")
	return nil
}

// StopRecording 停止 CPU 分析并写入堆栈信息
func (cpu *cpuTask) StopRecording() error {
	if err := cpu.CheckState(CPUStateFINISHED); err != nil {
		return err
	}
	pprof.StopCPUProfile()                                   //停止 CPU 分析
	if err := pprof.WriteHeapProfile(cpu.heap); err != nil { // 写入堆分析数据
		return err
	}
	cpu.state = CPUStateFINISHED
	cpu.log("stop recording")
	return nil
}

func (cpu *cpuTask) CloseFile() error {
	if err := cpu.CheckState(CPUStateIDLE); err != nil {
		return err
	}
	err := cpu.profile.Close()
	err2 := cpu.heap.Close()
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	cpu.state = CPUStateIDLE
	cpu.profile = nil
	cpu.heap = nil
	cpu.log("close file connect recording")
	return nil
}

func newCpu() *cpuTask {
	return &cpuTask{
		state: CPUStateIDLE,
	}
}

func Get() *cpuTask {
	return cpu
}

var (
	cpu          *cpuTask
	autoDuration time.Duration
)

func init() {
	cpu = newCpu()
	autoDuration = time.Minute * 10
	go wait()
}

func AutoDuration(d time.Duration) {
	autoDuration = d
}

func Auto() error {
	c := Get()
	c.log("recording ->", autoDuration.String())
	if err := c.CreateFile(fmt.Sprintf("pprof_auto_%s", autoDuration.String())); err != nil {
		return err
	}
	if err := c.StartRecording(); err != nil {
		return err
	}
	timer := time.NewTimer(autoDuration)
	<-timer.C
	if err := c.StopRecording(); err != nil {
		return err
	}
	return c.CloseFile()
}

func Manual() error {
	c := Get()
	switch c.state {
	case CPUStateIDLE:
		if err := c.CreateFile("pprof_manual"); err != nil {
			return err
		}
		if err := c.StartRecording(); err != nil {
			return err
		}
	case CPUStateACTIVE:
		if err := c.StopRecording(); err != nil {
			return err
		}
		if err := c.CloseFile(); err != nil {
			return err
		}
	}
	return nil
}

func wait() {
	sigs := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)
	for sig := range sigs {
		switch sig {
		case syscall.SIGUSR1:
			if err := Manual(); err != nil {
				fmt.Println(err.Error())
			}
		case syscall.SIGUSR2:
			if err := Auto(); err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}
