//Copyright 2017.7 wallestore<wallestore@hotmail.com>
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

package service

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

var (
	heart_start_time = time.Now().UTC()
)

type Monitor struct {
	CurrentServiceDate string //当前服务器时间
	ProgramRuntime     string //程序连续运行时长
	GoroutNum          int    //当前创建go程数量
}

/*
  http router add:
    (path: "/heartbeat", handle: {package}.HttpHeartbeat)
    (path: "/heartbeat", handle: server.HttpHeartbeat)
*/

func HttpHeartbeat(w http.ResponseWriter, r *http.Request) {
	buf, err := heartbeat()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Write(buf)
}

func heartbeat() ([]byte, error) {
	currentTime := time.Now()
	sctMsg := Monitor{
		CurrentServiceDate: currentTime.String(),
		ProgramRuntime:     time.Since(heart_start_time).String(),
		GoroutNum:          runtime.NumGoroutine(),
	}
	byteMsg, err := json.Marshal(sctMsg)
	if err != nil {
		return nil, err
	}
	return byteMsg, nil

}
