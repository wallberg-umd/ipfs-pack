package main

import (
	"fmt"
	"strings"

	human "github.com/dustin/go-humanize"
	net "github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	QClrLine = "\033[K"
	QReset   = "\033[2J"
)

/*
Move the cursor up N lines:
  \033[<N>A
- Move the cursor down N lines:
  \033[<N>B
- Move the cursor forward N columns:
  \033[<N>C
- Move the cursor backward N columns:
  \033[<N>D
*/

const (
	Clear = 0
)

const (
	Black = 30 + iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	LightGray
)

const (
	LightBlue = 94
)

func color(color int, s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, s)
}

func padPrint(line int, label, value string) {
	putMessage(line, fmt.Sprintf("%s%s%s", label, strings.Repeat(" ", 20-len(label)), value))
}

func putMessage(line int, mes string) {
	fmt.Printf("\033[%d;0H%s%s", line, QClrLine, mes)
}

func printDataSharedLine(line int, bup uint64, totup int64, rateup float64) {
	pad := "            "
	a := fmt.Sprintf("%d            ", bup)[:12]
	b := (human.Bytes(uint64(totup)) + pad)[:12]
	c := (human.Bytes(uint64(rateup)) + "/s" + pad)[:12]

	padPrint(line, "", a+b+c)
}

type Log struct {
	Size      int
	StartLine int
	Messages  []string
	beg       int
	end       int
}

func NewLog(line, size int) *Log {
	return &Log{
		Size:      size,
		StartLine: line,
		Messages:  make([]string, size),
		end:       -1,
	}
}

func (l *Log) Add(m string) {
	l.end = (l.end + 1) % l.Size
	if l.Messages[l.end] != "" {
		l.beg++
	}
	l.Messages[l.end] = m
}

func (l *Log) Print() {
	for i := 0; i < l.Size; i++ {
		putMessage(l.StartLine+i, l.Messages[(l.beg+i)%l.Size])
	}
}

type LogNotifee struct {
	addMes chan<- string
}

func (ln *LogNotifee) Listen(net.Network, ma.Multiaddr)      {}
func (ln *LogNotifee) ListenClose(net.Network, ma.Multiaddr) {}
func (ln *LogNotifee) Connected(_ net.Network, c net.Conn) {
	ln.addMes <- fmt.Sprintf("New connection from %s", c.RemotePeer().Pretty())
}

func (ln *LogNotifee) Disconnected(_ net.Network, c net.Conn) {
	ln.addMes <- fmt.Sprintf("Lost connection to %s", c.RemotePeer().Pretty())
}

func (ln *LogNotifee) OpenedStream(net.Network, net.Stream) {}
func (ln *LogNotifee) ClosedStream(net.Network, net.Stream) {}
