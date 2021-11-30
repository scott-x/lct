/*
* @Author: scottxiong
* @Date:   2021-11-22 09:45:12
* @Last Modified by:   scottxiong
* @Last Modified time: 2021-11-30 21:20:42
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

//just look for one item
type Location struct {
	Folders    []string          //folders: 可能存在的路径
	ExpectT    int               //0 file 1 folder 2 mix(file + folder)
	Re         string            //底层会被封装成regexp
	IgnoreFunc func(string) bool //参数是所有的dir
	result     string            //最终匹配结果: 有可能为空（match放全局如果有多个实例会污染变量）
	workers    int               //current workers
	maxWorkers int               //max workers
	ch_task    chan string
	ch_done    chan bool
	ch_matched chan string
	ch_result  chan string
	mutex      *sync.RWMutex
}

//find all
type Location2 struct {
	Folders    []string          //folders: 可能存在的路径
	ExpectT    int               //0 file 1 folder
	Re         string            //将来会被封装成regexp
	IgnoreFunc func(string) bool //参数是所有的dir
	Do         func(string)      //how to deal with the matched item
	workers    int               //current workers
	maxWorkers int               //max workers
	ch_task    chan string
	ch_done    chan bool
	ch_matched chan string
	ch_result  chan string
	mutex      *sync.RWMutex
}

func (l2 *Location2) Locate() {
	l2.mutex = new(sync.RWMutex)
	l2.workers = len(l2.Folders) //初始化worker
	l2.maxWorkers = 1 << 5
	l2.ch_task = make(chan string, 64)
	l2.ch_done = make(chan bool)
	l2.ch_matched = make(chan string)
	l2.ch_result = make(chan string)

	for _, folder := range l2.Folders {
		go walk2(folder, true, *l2)
	}

	wait2(l2)
}

//look for one item
func (l *Location) Locate() (string, time.Duration) {
	//init
	l.mutex = new(sync.RWMutex)
	l.workers = len(l.Folders) //初始化worker
	l.maxWorkers = 1 << 5
	l.ch_task = make(chan string, 64)
	l.ch_done = make(chan bool)
	l.ch_matched = make(chan string)
	l.ch_result = make(chan string)

	t1 := time.Now()

	for _, folder := range l.Folders {
		go walk1(folder, true, *l)
	}

	wait(l)
	return l.result, time.Since(t1)
}

func wait(l *Location) {
	for {
		select {
		case task := <-l.ch_task:
			l.mutex.Lock()
			l.workers++
			l.mutex.Unlock()
			// log.Println(workers)
			go walk1(task, true, l)
		case <-l.ch_done:
			l.mutex.Lock()
			l.workers--
			l.mutex.Unlock()
			// log.Println(workers)
			if l.workers == 0 {
				return
			}
		case result := <-l.ch_matched:
			l.result = result
			return
		}
	}
}

func wait2(l2 *Location2) {
	for {
		select {
		case task := <-l2.ch_task:
			l2.mutex.Lock()
			l2.workers++
			l2.mutex.Unlock()
			// log.Println(workers)
			go walk2(task, true, l2)
		case <-l2.ch_done:
			l2.mutex.Lock()
			l2.workers--
			flag := l2.workers == 0
			l2.mutex.Unlock()
			// log.Println(workers)
			if flag {
				return
			}

		case item := <-l2.ch_result:
			l2.Do(item)
		case result := <-l2.ch_matched:
			if l2.Do == nil {
				panic("Location.Do(string) must be implement!")
			}

			go func() {
				l2.ch_result <- result
			}()

			if l2.ExpectT == 2 {
				fi, _ := os.Stat(result)
				if fi.IsDir() {
					l2.mutex.Lock()
					flag := l2.workers < l2.maxWorkers
					l2.mutex.Unlock()
					if flag {
						l2.mutex.Lock()
						l2.workers++
						l2.mutex.Unlock()
						go walk2(result, true, l2)
					} else {
						walk2(result, false, l2)
					}
				}
			}
		}
	}
}

func walk1(dir string, goroutine bool, l *Location) {
	__re := regexp.MustCompile(l.Re)

	fls, _ := ioutil.ReadDir(dir)

	for _, v := range fls {
		name := v.Name()
		if v.IsDir() {
			new_dir := path.Join(dir, name)

			//ignore
			if l.IgnoreFunc != nil {
				if l.IgnoreFunc(new_dir) {
					continue
				}
			}

			//T==0,1,2 都可能会走这里
			//T==0 直接分配任务
			if l.ExpectT == 0 {
				l.mutex.Lock()
				flag := l.workers < l.maxWorkers
				l.mutex.Unlock()
				if flag {
					l.ch_task <- new_dir
				} else {
					walk1(new_dir, false, l)
				}
			} else {
				//T==1,2
				result := __re.FindString(name)

				if len(result) > 0 {
					//found
					l.ch_matched <- new_dir
				} else {
					l.mutex.Lock()
					flag := l.workers < l.maxWorkers
					l.mutex.Unlock()
					if flag {
						l.ch_task <- new_dir
					} else {
						walk1(new_dir, false, l)
					}
				}
			}
		} else {

			if l.ExpectT == 1 {
				continue
			}

			//file
			result := __re.FindString(name)

			//T==0或2都要放行
			if len(result) > 0 {
				//found
				l.ch_matched <- path.Join(dir, name)
			} else {
				continue
			}
		}
	}

	if goroutine {
		l.ch_done <- true
	}
}

func walk2(dir string, goroutine bool, l *Location2) {
	__re := regexp.MustCompile(l.Re)

	fls, _ := ioutil.ReadDir(dir)

	for _, v := range fls {
		name := v.Name()
		if v.IsDir() {
			new_dir := path.Join(dir, name)

			//ignore
			if l.IgnoreFunc != nil {
				if l.IgnoreFunc(new_dir) {
					continue
				}
			}

			//T==0,1,2 都可能会走这里
			//T==0 直接分配任务
			if l.ExpectT == 0 {
				l.mutex.Lock()
				flag := l.workers < l.maxWorkers
				l.mutex.Unlock()
				if flag {
					l.ch_task <- new_dir
				} else {
					walk2(new_dir, false, l)
				}
			} else {
				//T==1,2
				result := __re.FindString(name)

				if len(result) > 0 {
					//found
					l.ch_matched <- new_dir
				} else {
					l.mutex.Lock()
					flag := l.workers < l.maxWorkers
					l.mutex.Unlock()
					if flag {
						l.ch_task <- new_dir
					} else {
						walk2(new_dir, false, l)
					}
				}
			}
		} else {

			if l.ExpectT == 1 {
				continue
			}

			//file
			result := __re.FindString(name)

			//T==0或2都要放行
			if len(result) > 0 {
				//found
				l.ch_matched <- path.Join(dir, name)
			} else {
				continue
			}
		}
	}

	if goroutine {
		l.ch_done <- true
	}
}
