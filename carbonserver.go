package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/akamensky/argparse"
)

var running int = 1
var verbose bool

var fileStat string
var fileDetail string

type Counters struct {
	mx sync.Mutex
	m  map[string]int64
}

func NewCounters() *Counters {
	return &Counters{
		m: make(map[string]int64),
	}
}

func (c *Counters) Inc(key string) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.m[key]++
}

func (c *Counters) sortedKeys() []string {
	keys := make([]string, len(c.m))
	i := 0
	for k := range c.m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

var metrics = NewCounters()

func main() {
	parser := argparse.NewParser("carbonserver", "Carbonserver for testing")

	host := parser.String("a", "address", &argparse.Options{Required: false, Help: "Listen address", Default: "127.0.0.1"})
	port := parser.String("p", "port", &argparse.Options{Required: false, Help: "Listen port", Default: "2003"})
	verb := parser.Flag("v", "verbose", &argparse.Options{Help: "Enable verbose mode"})
	stat := parser.String("s", "stat", &argparse.Options{Required: false, Help: "file with metrics by key count", Default: ""})
	detail := parser.String("d", "detail", &argparse.Options{Required: false, Help: "file with received metrics", Default: ""})

	err := parser.Parse(os.Args)
	exit_on_error(err)

	fileStat = *stat
	fileDetail = *detail
	verbose = *verb

	addr, err := net.ResolveTCPAddr("tcp", *host+":"+*port)
	exit_on_error(err)

	listener, err := net.ListenTCP("tcp", addr)
	exit_on_error(err)

	fmt.Printf("Listening on %s:%s\n", *host, *port)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		running = 0
		time.Sleep(100 * time.Millisecond)
		statistic()
		os.Exit(1)
	}()

	cDetail := make(chan string, 1000)
	if fileDetail != "" {
		go func() {
			var w *bufio.Writer
			var file *os.File
			file, err = os.Create(fileDetail)
			exit_on_error(err)
			w = bufio.NewWriter(file)
			defer file.Close()
			for {
				select {
				case msg := <-cDetail:
					fmt.Fprint(w, msg)
				}
			}
		}()
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		} else {
			go client(conn, cDetail)
		}
	}
}

func client(conn net.Conn, c chan<- string) {
	defer conn.Close()

	fmt.Printf("Connected to: %s\n", conn.RemoteAddr().String())

	b := bufio.NewReader(conn)
	for running == 1 {
		line, err := b.ReadBytes('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Printf("Error reading: %s\n", err.Error())
			continue
		}
		if running == 0 {
			break
		}
		if fileDetail != "" {
			c <- string(line)
		}
		s := strings.Split(string(line), " ")
		if len(s) != 3 {
			fmt.Printf("Malformed: %s\n", string(line))
			continue
		}
		metrics.Inc(s[0])

	}
}

func statistic() {
	var count int64
	var w *bufio.Writer
	var file *os.File
	var err error
	keys := metrics.sortedKeys()
	if fileStat != "" {
		file, err = os.Create(fileStat)
		exit_on_error(err)
		w = bufio.NewWriter(file)
		defer file.Close()
	}

	for k := range keys {
		key := keys[k]
		value := metrics.m[key]
		count += value
		if fileStat == "" {
			fmt.Printf("%s %d\n", key, value)
		} else {
			fmt.Fprintf(w, "%s %d\n", key, value)
		}
	}
	if fileStat == "" {
		fmt.Printf("total %d\n", count)
	} else {
		fmt.Fprintf(w, "total %d\n", count)
	}
	w.Flush()
}

func exit_on_error(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
