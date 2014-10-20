package glogger

import (
	"github.com/golang/glog"
	"log"
)

type Glogger struct {
	*log.Logger
}

func New() *Glogger {
	l := log.New(nil, "", 0)
	return &Glogger{l}
}

func (l *Glogger) Fatal(v ...interface{}) {
	glog.Fatal(v)
}

func (l *Glogger) Fatalf(format string, v ...interface{}) {
	glog.Fatalf(format, v)
}

func (l *Glogger) Fatalln(v ...interface{}) {
	glog.Fatalln(v)
}

func (l *Glogger) Flags() int {
	return log.LstdFlags
}

func (l *Glogger) Output(calldepth int, s string) error {
	return nil
}

func (l *Glogger) Panic(v ...interface{}) {
	glog.Error(v)
}

func (l *Glogger) Panicf(format string, v ...interface{}) {
	glog.Error(format, v)
}

func (l *Glogger) Panicln(v ...interface{}) {
	glog.Error(v)
}

func (l *Glogger) Prefix() string {
	return ""
}

func (l *Glogger) Print(v ...interface{}) {
	glog.Info(v)
}

func (l *Glogger) Printf(format string, v ...interface{}) {
	glog.Infof(format, v)
}

func (l *Glogger) Println(v ...interface{}) {
	glog.Infoln(v)
}

func (l *Glogger) SetFlags(flag int) {

}

func (l *Glogger) SetPrefix(prefix string) {

}
