package util

import (
    "fmt"
    "runtime"
)

func HandlePanic(err error, function string) {
    if r:= recover();r != nil {
       buf := make([]byte, 10000)
       runtime.Stack(buf, false)
       if err != nil {
           *err = fmt.Errorf("%v", r)
       }
    }
}
