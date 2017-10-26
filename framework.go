//Copyright 2017 wallestore<wallestore@hotmail.com>
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	verbose_mode bool = false
	iframe       *Framework
)

type Framework struct {
	AppName         string                     //service name
	Config          interface{}                //config struct
	On_start_once   []func()                   //on_start_once func list
	On_time_loop    map[time.Duration][]func() //time_loop func list
	On_stop_once    []func()                   //on_stop_once func list
	Time_loop_close chan bool                  //time
}

var errorLog = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
var printLog = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

func init() {
	iframe = New()
}

//Create new instance
func New() *Framework {
	hook := new(Framework)
	hook.AppName = "app_name"
	hook.Config = nil
	hook.On_start_once = []func(){}
	hook.On_stop_once = []func(){}
	hook.On_time_loop = map[time.Duration][]func(){}
	hook.Time_loop_close = make(chan bool)
	return hook
}

//Get framework global pointer
func GetFramework() *Framework {
	return iframe
}

//set app name
func SetAppName(name string) { iframe.SetAppName(name) }
func (iframe *Framework) SetAppName(name string) {
	iframe.AppName = name
}

//set config
func SetConfig(conf interface{}) { iframe.SetConfig(conf) }
func (iframe *Framework) SetConfig(conf interface{}) {
	iframe.Config = conf
}

//get config
func GetConfig() interface{} { return iframe.Config }

//close timeloop
func CloseTimeLoop() { iframe.CloseTimeLoop() }
func (iframe *Framework) CloseTimeLoop() {
	close(iframe.Time_loop_close)
}

//start application
func Boot() {
	iframe.Start()
	iframe.Loop()
}
func (iframe *Framework) Boot() {

	iframe.onStartOnceLoop()

	//heartbeat func
	go iframe.timeLoop()

}

//start application
func Start() {
	iframe.Start()
	iframe.Loop()
}
func (iframe *Framework) Start() {

	iframe.onStartOnceLoop()
	// verbose monitor
	go signalAction()
	//loop func
	go iframe.timeLoop()
}

//stop application
func Stop() { iframe.Stop() }
func (iframe *Framework) Stop() {
	for _, fu := range iframe.On_stop_once {
		fu()
	}
}

//loop
func Loop() { iframe.Loop() }
func (iframe *Framework) Loop() {
	pid := os.Getpid()
	log.Println("PID", pid)
	// signal listen
	ch := make(chan os.Signal)
	//
	signal.Notify(ch, syscall.SIGSTOP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGKILL)
	c := <-ch

	//close all service
	switch c {
	case syscall.SIGSTOP, syscall.SIGQUIT, syscall.SIGTERM, os.Interrupt: //kill -19
		fmt.Println("stop ....", iframe.AppName)
		iframe.CloseTimeLoop()
	case syscall.SIGHUP, syscall.SIGKILL: // kill -1
		iframe.CloseTimeLoop()
		log.Println("sign up:", iframe.AppName)
	}
	iframe.Stop()
}

//add on_start_once func
func Init(fu func()) { iframe.Init(fu) }
func (self *Framework) Init(fu func()) {
	self.On_start_once = append(self.On_start_once, fu)
}

//add heartbeat func
func Heartbeat(t time.Duration, fu func()) { iframe.Heartbeat(t, fu) }
func (iframe *Framework) Heartbeat(t time.Duration, fu func()) {
	if _, ok := iframe.On_time_loop[t]; !ok {
		iframe.On_time_loop[t] = []func(){}
	}
	iframe.On_time_loop[t] = append(iframe.On_time_loop[t], fu)
}

//add stop func
func Exit(fu func()) { iframe.Exit(fu) }
func (iframe *Framework) Exit(fu func()) {
	iframe.On_stop_once = append(iframe.On_stop_once, fu)
}

//log
func Logln(v ...interface{}) { iframe.Logln(v) }
func (iframe *Framework) Logln(v ...interface{}) {
	//log.Output(2, fmt.Sprintln(v...))
	printLog.Output(2, fmt.Sprintln(v...))
}

//error log
func Errorln(v ...interface{}) { iframe.Errorln(v) }
func (iframe *Framework) Errorln(v ...interface{}) {
	errorLog.Output(2, fmt.Sprintln(v...))
}

// verbose log
func Verboseln(v ...interface{}) { iframe.VerboseLn(v) }
func (iframe *Framework) VerboseLn(v ...interface{}) {
	if verbose_mode {
		log.Output(2, fmt.Sprintln(v...))
	}
}

func signalAction() {
	ch := make(chan os.Signal, 1)
	for {
		signal.Notify(ch, syscall.SIGUSR1)
		c := <-ch
		switch c {
		case syscall.SIGUSR1:
			func() {
				verbose_mode = !verbose_mode
			}()
		}
	}
}

//==============================Internal Function=======================================================================
//parse json config
func ParseJsonConfig(path string, config interface{}) {
	file, _ := os.Open(path)
	defer file.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	json.Unmarshal(buf.Bytes(), &config)
}

//exec on_start_once func
func (iframe *Framework) onStartOnceLoop() {
	for _, fu := range iframe.On_start_once {
		fu()
	}
}

//exec time_loop func
func (iframe *Framework) timeLoop() {
	fmt.Println("ok")
	for t, funs := range iframe.On_time_loop {
		go func(t time.Duration, funs []func()) {
			tick := time.NewTicker(t)
			defer tick.Stop()
			var wait int64
			for {
				wait = time.Now().UnixNano()
				for _, fu := range funs {
					fu()
				}
				wait = (time.Now().UnixNano() - wait)
				if wait > t.Nanoseconds() {
					errorLog.Println("On time loop is too slow ! ", t, wait/int64(time.Millisecond))
				}
				select {
				case _, ok := <-tick.C:
					if !ok {
						return
					}
				case _, ok := <-iframe.Time_loop_close:
					if !ok {
						log.Println("Time loop stoped")

					}

					return
				}

			}
		}(t, funs)
	}
}
