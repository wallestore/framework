package main

import (
	"github.com/wallestore/framework"
	"time"
)

//init1
func init() {
	framework.SetAppName("heartbeat")
	framework.Heartbeat(10*time.Second, heartbeat)
}

//init2
//func init() {
//	iframe := framework.GetFramework()
//	iframe.AppName = "heartbeat"
//	iframe.Heartbeat(10*time.Second, heartbeat)
//}

func main() {
	framework.Start()
}

func heartbeat() {
	//fmt.Print("heartbeat ok")
	framework.Verboseln("heartbeat ok")
}
