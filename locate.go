/*
* @Author: scottxiong
* @Date:   2021-11-22 09:45:12
* @Last Modified by:   scottxiong
* @Last Modified time: 2021-11-22 10:59:31
 */
package lct

import (
	"io/ioutil"
	"path"
	"regexp"
	"time"
	"fmt"
	// "log"
)


var (
	workers    int = 0
	maxWorkers int = 32
	ch_task        = make(chan string, 0)
	ch_done        = make(chan bool, 0)
	ch_matched     = make(chan string, 0)
	match string
)

type Location struct {
	Folders  []string //folders: 可能存在的路径
	ExpectT int    //0 file 1 folder
	Re      string //将来会被封装成regexp
}

func (l *Location) Locate() (string, time.Duration){
	t1 := time.Now()

	for _, folder := range l.Folders {
		go walk(folder, true, l.ExpectT, l.Re)
	}

	wait(l)
	return match,time.Since(t1)
}

func wait(l *Location){
	for {
		select {
		case task:=<- ch_task:
			workers++
			// log.Println(workers)
			go walk(task, true, l.ExpectT,l.Re)
		case <-ch_done:
			workers--
			// log.Println(workers)
			if workers ==0 {
				return 
			}
		case result:=<-ch_matched:
			 match=result
			 return
		}
	}
}

func walk(dir string, goroutine bool, T int, re string) {
	__re := regexp.MustCompile(re)
	fls, _ := ioutil.ReadDir(dir)
	for _, v := range fls {
		name := v.Name()
		if v.IsDir() {
			result := __re.FindString(name)
			new_arr := path.Join(dir, name)
			if len(result) > 0 && T==1{
				//found
				ch_matched <- new_arr
			} else {
				if workers < maxWorkers {
					ch_task <- new_arr
				} else {
					walk(new_arr, false, T,re)
				}
			}

		}else {
			//file
			result := __re.FindString(name)
			if len(result) > 0 && T == 0{
				//found
				ch_matched <- path.Join(dir, name)
			}else{
				continue
			}
		}
	}

	if goroutine {
		ch_done <- true
	}
}
