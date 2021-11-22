/*
* @Author: scottxiong
* @Date:   2021-11-22 09:45:12
* @Last Modified by:   scottxiong
* @Last Modified time: 2021-11-22 10:59:31
 */
package lct

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sync"
	"time"
)

var (
	workers    int = 0
	maxWorkers int = 1 << 5
	ch_task        = make(chan string, 0)
	ch_done        = make(chan bool, 0)
	ch_matched     = make(chan string, 0)
	ch_result      = make(chan string, 0)
	match      string
	mutex      = &sync.RWMutex{}
)

//just look for one item
type Location struct {
	Folders    []string //folders: 可能存在的路径
	ExpectT    int      //0 file 1 folder 2 mix(file + folder)
	Re         string   //底层会被封装成regexp
	IgnoreFunc func(string) bool
}

//find all
type Location2 struct {
	Folders    []string //folders: 可能存在的路径
	ExpectT    int      //0 file 1 folder
	Re         string   //将来会被封装成regexp
	IgnoreFunc func(string) bool
	Do         func(string) //how to deal with the matched item
}

func (l2 *Location2) Locate() {
	mutex.Lock()
	workers = len(l2.Folders) //初始化worker
	mutex.Unlock()

	for _, folder := range l2.Folders {
		go walk(folder, true, l2.ExpectT, l2.Re, l2.IgnoreFunc)
	}

	wait2(l2)
}

//look for one item
func (l *Location) Locate() (string, time.Duration) {
	t1 := time.Now()

	for _, folder := range l.Folders {
		go walk(folder, true, l.ExpectT, l.Re, l.IgnoreFunc)
	}

	wait(l)
	return match, time.Since(t1)
}

func wait(l *Location) {
	for {
		select {
		case task := <-ch_task:
			mutex.Lock()
			workers++
			mutex.Unlock()
			// log.Println(workers)
			go walk(task, true, l.ExpectT, l.Re, l.IgnoreFunc)
		case <-ch_done:
			mutex.Lock()
			workers--
			mutex.Unlock()
			// log.Println(workers)
			if workers == 0 {
				return
			}
		case result := <-ch_matched:
			match = result
			return
		}
	}
}

func wait2(l *Location2) {
	for {
		select {
		case task := <-ch_task:
			mutex.Lock()
			workers++
			mutex.Unlock()
			// log.Println(workers)
			go walk(task, true, l.ExpectT, l.Re, l.IgnoreFunc)
		case <-ch_done:
			mutex.Lock()
			workers--
			flag := workers == 0
			mutex.Unlock()
			// log.Println(workers)
			if flag {
				return
			}

		case item := <-ch_result:
			l.Do(item)
		case result := <-ch_matched:
			if l.Do == nil {
				panic("Location.Do(string) must be implement!")
			}

			go func() {
				ch_result <- result
			}()

			if l.ExpectT == 2 {
				fi, _ := os.Stat(result)
				if fi.IsDir() {
					mutex.Lock()
					flag := workers < maxWorkers
					mutex.Unlock()
					if flag {
						mutex.Lock()
						workers++
						mutex.Unlock()
						go walk(result, true, l.ExpectT, l.Re, l.IgnoreFunc)
					} else {
						walk(result, false, l.ExpectT, l.Re, l.IgnoreFunc)
					}
				}
			}
		}
	}
}

func walk(dir string, goroutine bool, T int, re string, fn func(fd string) bool) {
	__re := regexp.MustCompile(re)

	fls, _ := ioutil.ReadDir(dir)

	for _, v := range fls {
		name := v.Name()
		if v.IsDir() {
			new_dir := path.Join(dir, name)

			//ignore
			if fn != nil {
				if fn(fd) {
					continue
				}
			}

			//T==0,1,2 都可能会走这里
			//T==0 直接分配任务
			if T == 0 {
				mutex.Lock()
				flag := workers < maxWorkers
				mutex.Unlock()
				if flag {
					ch_task <- new_dir
				} else {
					walk(new_dir, false, T, re, fn)
				}
			} else {
				//T==1,2
				result := __re.FindString(name)

				if len(result) > 0 {
					//found
					ch_matched <- new_dir
				} else {
					mutex.Lock()
					flag := workers < maxWorkers
					mutex.Unlock()
					if flag {
						ch_task <- new_dir
					} else {
						walk(new_dir, false, T, re, fn)
					}
				}
			}
		} else {

			if T == 1 {
				continue
			}

			//file
			result := __re.FindString(name)

			//T==0或2都要放行
			if len(result) > 0 {
				//found
				ch_matched <- path.Join(dir, name)
			} else {
				continue
			}
		}
	}

	if goroutine {
		ch_done <- true
	}
}
