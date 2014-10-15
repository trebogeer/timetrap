package main

import (
	"bufio"
	"flag"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	//    "fmt"
)

var (
	host      = flag.String("host", "localhost", "MongoDB host to connect to.")
	port      = flag.Int("port", 27017, "MongoDB port to connect to.")
	user      = flag.String("u", "midori", "MongoDB username.")
	pwd       = flag.String("p", "midori", "MongoDB password.")
	audb      = flag.String("authdb", "admin", "MongoDB database to authenticate against.")
	dbName    = flag.String("db", "midori", "MongoDB database to store metrics to.")
	collName  = flag.String("c", "mstat", "MongoDB collection to store metrics to.")
	cphost    = flag.Bool("cph", true, "Store each host's metrics to a separate collection.")
	dbg       = flag.Bool("dbg", false, "Print more output during execution if true.")
	logfile   = flag.String("logfile", os.TempDir()+string(os.PathSeparator)+"mstat.log", "Log file path.")
	mongostat = flag.String("mongostat", "mongostat", "mongostat executable path.")
	muname    = flag.String("muname", "DBMON", "Mongostat username.")
	mpass     = flag.String("mpassword", "ch3ck1ng", "Mongostat password.")
	hostport  = flag.String("hostport", "localhost:27017", "Host and Port for mongostat to connect foramtted as 'localhost:27017'.")
	mauthdb   = flag.String("mauthdb", "admin", "DB for mongodtat to authenticate agaist.")
	interval  = flag.String("interval", "1", "Interval in seconds to poll data for mongostat.")
)

func main() {

	flag.Parse()
	f, err := createLogFile(*logfile)
	//    fmt.Println("Created log file.")
	if err == nil {
		log.SetOutput(f)
		defer f.Close()
	}
	log.Println("Initialized logger.")
	mdbDialInfo := &mgo.DialInfo{
		Addrs:    []string{*host + ":" + strconv.Itoa(*port)},
		Source:   *audb,
		Username: *user,
		Password: *pwd,
		Timeout:  5 * time.Second,
	}

	log.Println("MongoDB Host: " + *host)
	log.Printf("MongoDB Port: %v", *port)
	log.Println("MongoDB User: " + *user)
	reg, _ := regexp.Compile(".")
	log.Println("MongoDB Password: " + reg.ReplaceAllString(*pwd, "*"))
	log.Println("MongoDB Auth Database: " + *audb)
	log.Println("MongoDB Database: " + *dbName)
	log.Println("MongoDB Collection: " + *collName)
	log.Printf("Collection per host: %v", *cphost)
	log.Printf("Debug: %v", *dbg)
	session, err := mgo.DialWithInfo(mdbDialInfo)
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	session.EnsureSafe(&mgo.Safe{W: 0, FSync: false})
	c := session.DB(*dbName).C(*collName)

	cm := make(map[string]*mgo.Collection)

	stdout, cmd, err := runMongostat(*muname, *mpass, *hostport, *mauthdb, *interval, *mongostat)
	if err != nil {
		log.Fatal("Failed to start mongostat: ", err)
	}
	defer stdout.Close()

	stripStars, _ := regexp.Compile("\\*")
	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	for err == nil {
		m := strings.Fields(stripStars.ReplaceAllString(line, ""))
		if len(m) == 22 {
			e12 := strings.Split(m[12], ":")
			e12_0 := e12[0]
			e12_1 := e12[1]
			var lp float64
			if len(e12_1) > 0 {
				lp = toFloat(e12_1[:len(e12_1)-1])
			} else {
				lp = 0
			}
			h_ := strings.Split(m[0], ":")[0]
			rep := m[20]
			dt := time.Now()
			sec := dt.Second()
			sec_s := strconv.Itoa(sec)
			id := mstatObjectId(h_, rep, time.Now())

			var coll *mgo.Collection
			if *cphost {
				key := h_ + rep
				hc, ok := cm[key]
				if !ok {
					hc_ := session.DB(*dbName).C(key)
					cm[key] = hc_
					coll = hc_
				} else {
					coll = hc
				}
			} else {
				coll = c
			}
			doc := bson.M{
				"h": m[0], "i": toInt(m[1]), "q": toInt(m[2]), "u": toInt(m[3]), "d": toInt(m[4]), "g": toInt(m[5]),
				"c": toInt(m[6]), "f": toInt(m[7]), "m": m[8], "v": m[9], "r": m[10], "pf": toInt(m[11]), "ldb": e12_0,
				"lp": lp, "im": toInt(m[13]), "rq": toInt(strings.Split(m[14],
					"|")[0]), "wq": toInt(strings.Split(m[14], "|")[1]), "ar": toInt(strings.Split(m[15],
					"|")[0]), "aw": toInt(strings.Split(m[15], "|")[1]), "ni": m[16], "no": m[17], "cn": toInt(m[18]),
				"s": m[19], "repl": m[20], "t": m[21], "ts": dt}
			debug("Document: %v", doc)

			go func(doc bson.M, coll *mgo.Collection) {
				_, dberr := coll.Upsert(bson.M{"_id": id}, bson.M{"$set": bson.M{sec_s: doc}})
				if dberr != nil {
					log.Println(dberr)
				}
			}(doc, coll)
		} else {
			debug("Skipping line [%v] due to non-standard format:", line)
		}
		line, err = reader.ReadString('\n')
	}
	if err != io.EOF {
		log.Println(err)
	}
	if err = cmd.Wait(); err != nil {
		log.Println(err)
	}
	log.Println("Bye!")
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
		log.Printf(t+"\n", i)
	}

}

func createLogFile(logfile string) (*os.File, error) {
	/*	if _, err := os.Stat(logfile); !os.IsNotExist(err) {
			err = os.Rename(logfile, logfile+"."+time.Now().Format("2000-01-29T20-20-39.000"))
			if err != nil {
				return nil, err
			}
		}
	*/
	if f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return f, nil
	} else {
		return nil, err
	}
}

func runMongostat(uname, pass, hostport, authDB, interval, mongostat string) (io.ReadCloser, *exec.Cmd, error) {

	cmd := exec.Command(mongostat, "--host", hostport, "--username", uname,
		"--password", pass, "--authenticationDatabase", authDB, "--discover", interval)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, nil, err
	}
	return stdout, cmd, nil
}
