package main

import (
  "log"
  "time"
  "strconv"
)


func main() {
  d := time.Now()
  c := d.Truncate(time.Duration(1)*time.Second)

  log.Println(c)
  log.Println(strconv.FormatInt(c.UnixNano()/int64(time.Millisecond), 16))
}
