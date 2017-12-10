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
