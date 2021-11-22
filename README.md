# lct

一款高性能的文件/文件夹查找工具，只返回第一个找到的，一旦有结果，程序立即退出

核心struct

```go
type Location struct {
	Folders  []string //folders: 可能存在的路径
	ExpectT int    //0 file 1 folder
	Re      string //将来会被封装成regexp
}
```

### API 

- `func (l *Location) Locate() (string, time.Duration)`:返回结果和时间

### Example

```go
package main

import (
	"github.com/scott-x/lct"
)

func main() {
	l := &lct.Location {
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
		Re: "^U211042.*",
	}

	res, t:= l.Locate()
	fmt.Println(res,t)
}
```