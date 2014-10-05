package main

import (
	"math/rand"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (

	host = flag.String("host", "localhost", "MongoDB host to connect to.")
	port = flag.Int("port", 27017, "MongoDB port to connect to.")
	user = flag.String("u", "midori", "MongoDB username.")
	pwd = flag.String("p", "midori", "MongoDB password.")
	audb = flag.String("authdb", "admin", "MongoDB database to authenticate against.")
	dbName = flag.String("db", "midori", "MongoDB database to store metrics to.")
	collName = flag.String("c", "mstat", "MongoDB collection to store metrics to.")
	cphost = flag.Bool("cph", true, "Store each host's metrics to a separate collection.")
	dbg = flag.Bool("dbg", false, "Print more output during execution if true.")

)

func main() {

	flag.Parse()

	mdbDialInfo := &mgo.DialInfo{
		Addrs:    []string{*host + ":" + strconv.Itoa(*port)},
		Source:   *audb,
		Username: *user,
		Password: *pwd,
	}

	fmt.Println("MongoDB Host: " + *host)
	p := fmt.Sprintf("MongoDB Port: %v", *port)
	fmt.Println(p)
	fmt.Println("MongoDB User: " + *user)
	reg, _ := regexp.Compile(".*")
	fmt.Println("MongoDB Password: " + reg.ReplaceAllString(*pwd, "*"))
	fmt.Println("MongoDB Auth Database: " + *audb)
	fmt.Println("MongoDB Database: " + *dbName)
	fmt.Println("MongoDB Collection: " + *collName)
	fmt.Printf("Collection per host: %v\n", *cphost)
    fmt.Printf("Debug: %v\n", *dbg)
	session, err := mgo.DialWithInfo(mdbDialInfo)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	session.EnsureSafe(&mgo.Safe{W: 0, FSync: false})

	h_,err := os.Hostname()
    if err != nil {
        h_ = "localhost"
    }

    coll := session.DB(*dbName).C(h_)
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
	y := time.Now().Year()
    dt := time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
    
    for i:= 0; i < 60*60*24*365; i++ {
			
			sec := dt.Second()
			sec_s := strconv.Itoa(sec)
            id := mstatObjectId(h_, "test", dt)

             doc := bson.M{
						"h": h_, "i": r.Int()%1000, "q": r.Int()%1000, "u": r.Int()%1000, "d": r.Int()%1000, "g": r.Int()%1000,
						"c": r.Int()%1000, "f": r.Int()%1000, "m": r.Float32(), "v": r.Float32(), "r": r.Float32(), 
                        "pf": r.Int()%1100, "ldb": r.Float32(),
                        "lp": r.Float32(), "im": r.Int(), "rq": r.Int()%300, "wq": r.Int()%300, "ar": r.Int()%2000,
                        "aw": r.Int()%100, "ni": r.Float32(), "no": r.Float32(), "cn": r.Int()%500,
						"s": r.Float32(), "repl": r.Float32(), "t": r.Float32(), "ts": dt}
             debug("Document: %v", doc)

			_, dberr := coll.Upsert(bson.M{"_id": id}, bson.M{"$set": bson.M{sec_s: doc}})
			if dberr != nil {
				fmt.Println(dberr)
			}
            dt = dt.Add(time.Second)
	}
	fmt.Println("Bye!")
}

func toInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	} else {
		return i
	}
}

func toFloat(s string) float64 {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	} else {
		return i
	}
}

func mstatObjectId(host string, repl string, t time.Time) string {
	c := t.Truncate(time.Duration(1) * time.Minute)
	timeHex := strconv.FormatInt(c.UnixNano()/int64(time.Second), 16)
	return timeHex + host + repl
}

func debug(t string, i ...interface{}) {
  if *dbg {
     fmt.Printf(t+"\n", i)
  }

}
