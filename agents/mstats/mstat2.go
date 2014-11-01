package main

import (
	"bufio"
	"flag"
	"github.com/golang/glog"
	glogger "github.com/trebogeer/timetrap/glogger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
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
	mongostat = flag.String("mongostat", "mongostat", "mongostat executable path.")
	muname    = flag.String("muname", "midori", "Mongostat username.")
	mpass     = flag.String("mpassword", "midori", "Mongostat password.")
	hostport  = flag.String("hostport", "localhost:27017", "Host and Port for mongostat to connect foramtted as 'localhost:27017'.")
	mauthdb   = flag.String("mauthdb", "admin", "DB for mongodtat to authenticate agaist.")
	interval  = flag.String("interval", "1", "Interval in seconds to poll data for mongostat.")
)

func main() {

	glogger := glogger.New()
	mgo.SetLogger(glogger)

	flag.Parse()

	glog.V(1).Info("MongoDB Host: " + *host)
	glog.V(1).Infof("MongoDB Port: %v", *port)
	glog.V(1).Info("MongoDB User: " + *user)
	reg, _ := regexp.Compile(".")
	glog.V(1).Info("MongoDB Password: " + reg.ReplaceAllString(*pwd, "*"))
	glog.V(1).Info("MongoDB Auth Database: " + *audb)
	glog.V(1).Info("MongoDB Database: " + *dbName)
	glog.V(1).Info("MongoDB Collection: " + *collName)
	glog.V(1).Infof("Collection per host: %v", *cphost)

	mdbDialInfo := &mgo.DialInfo{
		Addrs:    []string{*host + ":" + strconv.Itoa(*port)},
		Source:   *audb,
		Username: *user,
		Password: *pwd,
		Timeout:  5 * time.Second,
	}

	session, err := mgo.DialWithInfo(mdbDialInfo)
	if err != nil {
		glog.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	session.EnsureSafe(&mgo.Safe{W: 0, FSync: false})
	c := session.DB(*dbName).C(*collName)

	cm := make(map[string]*mgo.Collection)

	stdout, cmd, err := runMongostat(*muname, *mpass, *hostport, *mauthdb, *interval, *mongostat)
	if err != nil {
		glog.Fatal("Failed to start mongostat: ", err)
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
			glog.V(2).Infof("Document: %v", doc)

			go func(doc bson.M, coll *mgo.Collection) {
				_, dberr := coll.Upsert(bson.M{"_id": id}, bson.M{"$set": bson.M{sec_s: doc}})
				if dberr != nil {
					glog.Error(dberr)
				}
			}(doc, coll)
		} else {
			glog.V(1).Infof("Skipping line [%v] due to non-standard format:", line)
		}
		line, err = reader.ReadString('\n')
	}
	if err != io.EOF {
		glog.Error(err)
	}
	if err = cmd.Wait(); err != nil {
		glog.Error(err)
	}
	glog.Info("Bye!")
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
