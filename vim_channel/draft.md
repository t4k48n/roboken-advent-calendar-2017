% Vimを使ったソケット通信

# ソケット通信機能

Vim8.0でソケット通信機能が追加されました。ソケット通信を使うことでVimと、異なるプロセスの間で通信ができます。このアップデートで、以前は外部スクリプトを使うなどの面倒な方法で実現していた通信を、Vimの機能でできるようになります。

以下の説明は`+channel`機能付きでコンパイルされたVimでのみ動作します。持ってない人はKaoriya Vimを使うとよいと思います。

# チャンネル

Vimでソケット通信をするにはチャンネルという機能を使います。チャンネルを使う簡単なVim Scriptを次に示します。

```
function! Callback(handle, message)
    echo a:message
endfunction
let chan = ch_open("localhost:6868")
call ch_sendexpr(chan, "sent message", {"callback": "Callback"})
sleep(100)
if ch_status(chan) == "open"
  call ch_close(chan)
endif
```

`ch_open("localhost:6868")`で、ローカルのポート`6868`で接続待ちしているサーバに接続し、チャンネルを開きます。その後`ch_sendexpr(...)`で、接続したチャンネルに対してメッセージ`"sent message"`を送信し、その応答を`"Callback"`関数で非同期に処理しています。最後に`ch_status(...)`でチャンネルが開いているかどうかを確認し、開いているなら`ch_close(...)`でチャンネルを閉じます。

動作確認用にGo言語で書いたエコーバックサーバを置いておきます。

```
package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	l, e := net.Listen("tcp", ":6868")
	if e != nil {
		return
	}
	for {
		c, e := l.Accept()
		if e != nil {
			continue
		}
		var b [500]byte
		n, e := c.Read(b[:])
		if e != nil {
			break
		}
		fmt.Println(string(b[:n]))
		c.Write(b[:n])
		time.Sleep(100 * time.Millisecond)
		c.Close()
		break
	}
}
```

# チャンネルの使用例：バスの時刻表

Vimからサーバに現在時刻を送信し、受け取ったサーバが次のバスの時刻をVimに返し、Vimが表示するアプリケーションです。iizuka氏がすばらしいWEBアプリを作っていた（露骨な媚）ので、同じものを作りたかったけど劣化版しか作れませんでした（憤怒）

コードは最後尾にまとめて貼り付けてますので、興味があれば動かしてみてください。

## 記事まとめ

次はもっと余裕を持って書こうと思います。

## 時刻表ソースコード

```
=== client.vim ===
let s:delimiter = fnamemodify(".", ":p")[-1:]
let s:ext = ""
if has('win32')
  let s:ext = ".exe"
endif

let BusServer = expand("<sfile>:p:h") . s:delimiter . "server" .s:ext
command! BusInit let BusJob = job_start(BusServer)
command! BusQuit call job_stop(BusJob)
command! Bus call BusQuery()

function! BusAnnounce(handle, message)
  echo join(["次のバスは", a:message, "に到着します．"], " ")
  call ch_close(a:handle)
endfunction
function! BusQuery()
  let s:handle = ch_open("localhost:6868")
  call ch_sendexpr(s:handle, strftime("%H:%M"), {"callback": "BusAnnounce"})
endfunction
```

```
=== server.go ===
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	sch_dat, err := filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), "schedule.csv"))
	if err != nil {
		os.Exit(1)
	}
	sch, err := loadSchdule(sch_dat)
	if err != nil {
		os.Exit(1)
	}
	l, err := net.Listen("tcp", ":6868")
	if err != nil {
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go serve(conn, sch)
	}
}

func serve(conn net.Conn, sch Schedule) {
	defer conn.Close()
	handle, timestr, err := recieveMessage(conn)
	if err != nil {
		return
	}
	curr := parseTime(timestr)
	next := sch.findNext(curr)
	nextstr := next.String()
	fmt.Println(curr)
	fmt.Println(next)
	err = sendMessage(conn, handle, nextstr)
	if err != nil {
		return
	}
}

func recieveMessage(conn net.Conn) (float64, string, error) {
	var b [500]byte
	n, err := conn.Read(b[:])
	fmt.Println(string(b[:n]))
	if err != nil {
		return 0, "", errors.New("")
	}
	var v [2]interface{}
	err = json.Unmarshal(b[:n], &v)
	if err != nil {
		return 0, "", errors.New("")
	}
	h, s := v[0].(float64), v[1].(string)
	return h, s, nil
}

func sendMessage(conn net.Conn, handle float64, timestr string) error {
	var v [2]interface{}
	v[0] = handle
	v[1] = timestr
	b, err := json.Marshal(&v)
	if err != nil {
		return errors.New("")
	}
	_, err = conn.Write(b)
	if err != nil {
		return errors.New("")
	}
	time.Sleep(1 * time.Second)
	return nil
}

type Time struct {
	Hour   int
	Minute int
}

func parseTime(str string) Time {
	var h, m int
	fmt.Sscanf(str, "%d:%d", &h, &m)
	return Time{Hour: h, Minute: m}
}

func (t Time) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

func (t Time) After(a Time) bool {
	if t.Hour > a.Hour {
		return true
	}
	if t.Hour == a.Hour && t.Minute > a.Minute {
		return true
	}
	return false
}

type Schedule []Time

func loadSchdule(fname string) (Schedule, error) {
	file, err := os.Open(fname)
	if err != nil {
		return Schedule{}, errors.New("")
	}
	r := bufio.NewReader(file)
	var s Schedule
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}
		s = append(s, parseTime(string(line)))
	}
	return s, nil
}

func (sch Schedule) findNext(t Time) Time {
	next := sch[0]
	for i := len(sch) - 1; i > -1; i-- {
		if t.After(sch[i]) {
			break
		}
		next = sch[i]
	}
	return next
}
```

```
=== schedule.csv ===
05:53
06:00
06:11
06:21
06:30
06:36
06:46
06:51
07:02
07:04
07:08
07:14
07:19
07:24
07:29
07:35
07:38
07:46
07:49
07:56
08:04
08:13
08:25
08:45
08:50
08:52
08:55
09:05
09:15
09:20
09:30
09:40
09:50
10:00
10:10
10:20
10:30
10:40
10:50
11:00
11:10
11:20
11:35
11:50
12:05
12:20
12:35
12:50
13:00
13:10
13:20
13:35
13:50
14:05
14:20
14:35
14:50
15:05
15:20
15:30
15:40
15:50
16:00
16:10
16:20
16:30
16:40
16:50
17:00
17:10
17:25
17:40
17:45
17:50
17:52
18:00
18:10
18:15
18:20
18:40
18:45
18:55
18:57
19:10
19:20
19:30
19:45
19:50
20:10
20:20
20:30
20:45
21:04
21:15
21:35
22:00
22:30
22:45
```
