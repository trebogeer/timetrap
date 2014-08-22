package main

import (
    "fmt"
    "bufio"
    "os"
    "io"
    "flag"
    "regexp"
    "strings"
    "strconv"
    "time"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
)

func main() {

    host := flag.String("host", "localhost", "MongoDB host to connect to.")
    port := flag.Int("port", 27017, "MongoDB port to connect to.")
    user := flag.String("u","midori", "MongoDB username.")
    pwd := flag.String("p", "midori", "MongoDB password.")
    audb := flag.String("authdb", "admin", "MongoDB database to authenticate against.")
    dbName := flag.String("db", "midori", "MongoDB database to store metrics to.")
    collName := flag.String("c", "mstat", "MongoDB collection to store metrics to.")

    flag.Parse()

    mdbDialInfo := &mgo.DialInfo {
        Addrs: []string{*host + ":"+ strconv.Itoa(*port)},
        Source: *audb,
        Username: *user,
        Password: *pwd,
    }

    fmt.Println("MongoDB Host: " + *host)
    p := fmt.Sprintf("MongoDB Port: %v", *port)
    fmt.Println(p)
    fmt.Println("MongoDB User: " + *user)
    reg, _:= regexp.Compile(".*")
    fmt.Println("MongoDB Password: " + reg.ReplaceAllString(*pwd, "*"))
    fmt.Println("MongoDB Auth Database: " + *audb)
    fmt.Println("MongoDB Database: " + *dbName)
    fmt.Println("MongoDB Collection: " + *collName)
    session, err := mgo.DialWithInfo(mdbDialInfo)
    if err != nil {
	   panic(err)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)
    session.EnsureSafe(&mgo.Safe{W:0, FSync:false})
    c := session.DB(*dbName).C(*collName)


//    err = c.Insert(bson.M{"hey":"there"})
//    if err != nil {
//	   panic(err)
//    }
    stripStars,_ := regexp.Compile("\\*")
    reader := bufio.NewReader(os.Stdin)
    line, err := reader.ReadString('\n')
    for err == nil {
       m := strings.Fields(stripStars.ReplaceAllString(line, ""))
       if len(m) == 22 {
           e12 := strings.Split(m[12],":")
           e12_0 := e12[0]
           e12_1 := e12[1]
           dberr :=
           c.Insert(bson.M{"h":m[0],"i":toInt(m[1]),"q":toInt(m[2]),"u":toInt(m[3]),"d":toInt(m[4]),"g":toInt(m[5]),"c":toInt(m[6]),
           "f":toInt(m[7]),"m":m[8],"v":m[9],"r":m[10],"pf":toInt(m[11]),"ldb":e12_0,"lp":toFloat(e12_1[0:len(e12_1)-1]),"im":toInt(m[13]),"rq":toInt(strings.Split(m[14],
           "|")[0]),"wq":toInt(strings.Split(m[14],"|")[1]),"ar":toInt(strings.Split(m[15],
           "|")[0]),"aw":toInt(strings.Split(m[15],"|")[1]),"ni":m[16],"no":m[17],"cn":toInt(m[18]),"s":m[19],"repl":m[20],"t":m[21],"ts":time.Now()})
           if dberr != nil {
               fmt.Println(dberr)
           }
       }
       line, err = reader.ReadString('\n')
    }
    if err != io.EOF {
       fmt.Println(err)
    }
    fmt.Println("Bye!")
}

func toInt(s string) int {
    i, err := strconv.Atoi(s)
    if err != nil {
      return -1
    } else {
      return i
    }
}

func toFloat(s string) float64 {
    i, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return -1
    } else {
        return i
    }
}
