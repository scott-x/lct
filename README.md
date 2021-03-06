# lct

一款高性能的文件/文件夹查找工具，经过`--race`测试

2个核心struct

```go
//find one
type Location struct {
	Folders  []string //folders: 可能存在的路径
	ExpectT int    //0 file 1 folder
	Re      string //将来会被封装成regexp
	IgnoreFunc func(string) bool
	MaxWorkers int
}

//find all
type Location2 struct {
	Folders []string //folders: 可能存在的路径
	ExpectT int      //0 file 1 folder 2: mix(file & folder)
	Re      string   //将来会被封装成regexp
	IgnoreFunc func(string) bool
	Do func(string) //当找到匹配项时回自动执行
	MaxWorkers int
}
```

### API 

- `func (l *Location) Locate() (string, time.Duration)`:只返回第一个找到的，一旦有结果，程序立即退出，返回结果和时间
- `func (l2 *Location2) Locate()`:find all，一旦匹配自动执行`l2.Do()`
- `func GetLatestFileTime(path string) int64`: //获取文件修改时间 返回unix时间戳

### Example

```go
package main

import (
	"fmt"
	"github.com/scott-x/lct"
	"strings"
)

var ignores = []string{
	"test",
}

func main() {
	l := &lct.Location{
		Folders: []string{
			// "/Volumes/datavolumn_bmkserver_Pub/新做稿/未开始",
			// "/Volumes/datavolumn_bmkserver_Pub/新做稿/进行中",
			// "/Volumes/datavolumn_bmkserver_Pub/新做稿/已结束",
			// "/Volumes/datavolumn_bmkserver_Design/Proofing",
			// "/Volumes/datavolumn_bmkserver_Design/Other",
			// "/Volumes/datavolumn_bmkserver_Design/WMT-Canada",
			// "/Volumes/datavolumn_bmkserver_Design/WMT-USA",
			"/Volumes/datavolumn_bmkserver_Design/WMT-USA/2022/USA_HOL_2022",
			"/Volumes/datavolumn_bmkserver_Pub/新做稿/进行中",
		},
		ExpectT: 1,
		Re:      "^U211042.*",
		IgnoreFunc: func(dir string) bool {
			var flag bool
			//fmt.Println(dir) //print all dir
			for _, ignore := range ignores {
				lastSlash := strings.LastIndex(dir, "/")
				name := dir[lastSlash+1:]
				if ignore == name {
					flag = true
					fmt.Println("ignore:", dir)
					break
				}
			}
			return flag
		},
	}

	res, t := l.Locate()
	fmt.Println(res, t)
}
```

```go
package main

import (
	"fmt"
	"github.com/scott-x/lct"
	"strings"
)

var ignores = []string{
	"test",
}

func main() {
	l := &lct.Location2{
		Folders: []string{
			// "/Volumes/datavolumn_bmkserver_Pub/新做稿/未开始",
			// "/Volumes/datavolumn_bmkserver_Pub/新做稿/进行中",
			// "/Volumes/datavolumn_bmkserver_Pub/新做稿/已结束",
			// "/Volumes/datavolumn_bmkserver_Design/Proofing",
			// "/Volumes/datavolumn_bmkserver_Design/Other",
			// "/Volumes/datavolumn_bmkserver_Design/WMT-Canada",
			// "/Volumes/datavolumn_bmkserver_Design/WMT-USA",
			"/Volumes/datavolumn_bmkserver_Design/WMT-USA/2022/USA_HOL_2022",
			"/Volumes/datavolumn_bmkserver_Pub/新做稿/进行中",
		},
		ExpectT: 2,
		Re:      "^U211042.*",
		IgnoreFunc: func(dir string) bool {
			var flag bool
			//fmt.Println(dir) //print all dir
			for _, ignore := range ignores {
				lastSlash := strings.LastIndex(dir, "/")
				name := dir[lastSlash+1:]
				if ignore == name {
					flag = true
					fmt.Println("ignore:", dir)
					break
				}
			}
			return flag
		},
		Do: func(item string) {
			fmt.Println(item)
		},
	}

	l.Locate()
}
```