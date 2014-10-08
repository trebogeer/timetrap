package main

import (
	"bufio"
	"flag"
//	"fmt"
//	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"io"
    "os"
    "io/ioutil"
    //"log"
	//"os"
    "os/exec"
//    "bytes"
	"regexp"
	"strconv"
	//"strings"
	"time"
    "gopkg.in/yaml.v2"
    //stdlog "log"

    log "github.com/golang/glog"
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
    file_path = flag.String("fp", "/tmp/test", "Log file to scrape.")
    yaml_rules_file = flag.String("rf", "rules.yaml", "Regular experession rules ")

)

func main() {

	flag.Parse()

	log.V(1).Info("MongoDB Host: " + *host)
    log.V(1).Infof("MongoDB Port: %v", *port)
	log.V(1).Info("MongoDB User: " + *user)
	reg, _ := regexp.Compile(".*")
	log.V(1).Info("MongoDB Password: " + reg.ReplaceAllString(*pwd, "*******"))
	log.V(1).Info("MongoDB Auth Database: " + *audb)
	log.V(1).Info("MongoDB Database: " + *dbName)
	log.V(1).Info("MongoDB Collection: " + *collName)
	log.V(1).Info("Collection per host: %v", *cphost)
    log.V(1).Info("Log file: %v", *file_path)
    log.V(1).Info("Rules yaml file path: %v", *yaml_rules_file)

    // yaml rules processing
    rules := make(map[interface{}]interface{})
    yamlBytes, err := ioutil.ReadFile(*yaml_rules_file)
    if err != nil {
        log.Fatal("Error reading rules file:", err)
        os.Exit(1)
    }
    err = yaml.Unmarshal(yamlBytes, &rules)
    if err != nil {
        log.Fatal("Error parsing rules file:", err)
        os.Exit(1)
    }

    log.Info("%v", rules)
    // tailing output

    cmd := exec.Command("tail", "-F", *file_path)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatal("Error tailing file",err)
        os.Exit(1)
    }
    if err := cmd.Start();err != nil {
        //log.Printf("Error executing command: %v %v", tail_path, *file_path)
        log.Fatal("Error executing command: ", err)
    }

/*	mdbDialInfo := &mgo.DialInfo{
		Addrs:    []string{*host + ":" + strconv.Itoa(*port)},
		Source:   *audb,
		Username: *user,
		Password: *pwd,
	}
*/
/*session, err := mgo.DialWithInfo(mdbDialInfo)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	session.EnsureSafe(&mgo.Safe{W: 0, FSync: false})
	c := session.DB(*dbName).C(*collName)
    log.Printf("%v", c)
	cm := make(map[string]*mgo.Collection)
    log.Printf("%v", cm)
	stripStars, _ := regexp.Compile("\\*")
    log.Printf("%v", stripStars)*/
	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	for err == nil {

    log.Info(line)
	/*	m := strings.Fields(stripStars.ReplaceAllString(line, ""))
		if len(m) == 22 {
			e12 := strings.Split(m[12], ":")
			e12_0 := e12[0]
			e12_1 := e12[1]
			var lp float64
			if len(e12_1) > 0 {
				lp = toFloat(e12_1[:len(e12_1) - 1])
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

			_, dberr := coll.Upsert(bson.M{"_id": id}, bson.M{"$set": bson.M{sec_s: doc}})
			if dberr != nil {
				fmt.Println(dberr)
			}
		} else {
            debug("Skipping line [%v] due to non-standard format:", line)
        }*/
		line, err = reader.ReadString('\n')
	}
	if err != io.EOF {
		log.Error(err)
	}
    if err := cmd.Wait();err != nil {
        log.Fatal(err)
    }
	log.Info("Bye!")
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

