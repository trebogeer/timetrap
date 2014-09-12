package mongo

import (
	"errors"
	"log"
	"os"
	"regexp"
	"time"
	//"reflect"
	//env "github.com/kelseyhightower/envconfig"
	mgo "gopkg.in/mgo.v2"
	bson "gopkg.in/mgo.v2/bson"
	cnv "strconv"
)

var (
	masterSession session
	secs_per_doc  = 60
	zero_host     = "0000000000000000"
	Empty_Graph   = make(map[string]Points)
)

type (
	session struct {
		dialInfo *mgo.DialInfo
		ms       *mgo.Session
	}

	/*  mconf struct {
	        Host   string
	        Port   string
	        User   string
	        Pass   string
	        AuthDB string
	    }
	*/
)

type XY []interface{}
type Points []XY
type round_time func(t time.Time) time.Time

func Init(host, port, authdb, user, password string) error {

	if masterSession.ms != nil {
		return nil
	}

	//    mgo.SetDebug(true)

	var aLogger *log.Logger
	aLogger = log.New(os.Stderr, "", log.LstdFlags)
	mgo.SetLogger(aLogger)

	dialInfo := &mgo.DialInfo{
		Addrs:    []string{host + ":" + port},
		Source:   authdb,
		Username: user,
		Password: password,
	}

	ms, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return err
	}
	//    defer Client.Close()

	ms.SetMode(mgo.Monotonic, true)
	ms.EnsureSafe(&mgo.Safe{W: 0, FSync: false})
	ms.SetBatch(1000)
	ms.SetPrefetch(0.25)

	masterSession = session{dialInfo, ms}

	log.Println("Mongo client is initialized.")
	return nil
}

func Shutdown() {
	if masterSession.ms != nil {
		masterSession.ms.Close()
	}
	log.Println("MGO master session shutdown is complete.")
}

func GetGraphData(db, c, x, y string, from, to time.Time, labels []string) (error, map[string]Points) {
	//TODO validate input
	if len(db) == 0 || len(c) == 0 || len(x) == 0 || len(y) == 0 || len(labels) == 0 {
		return errors.New("Illegal argumet"), Empty_Graph
	}
	proj := bson.M{}
	for i := 0; i < secs_per_doc; i++ {
		sec_pref := cnv.Itoa(i) + "."
		proj[sec_pref+y] = 1
		proj[sec_pref+x] = 1
		for a := range labels {
			proj[sec_pref+labels[a]] = 1
		}
	}

	query := bson.M{"_id": bson.M{"$gte": objId(from, t_rup_min), "$lt": objId(to, t_rdown_min)}}

	s := masterSession.ms.Copy()
	coll := s.DB(db).C(c)
	defer s.Close()
	cur := coll.Find(query).Sort("_id").Select(proj).Iter()

	res := make(map[string]Points)
	entry := bson.M{}

	for cur.Next(&entry) {
		for s := 0; s < secs_per_doc; s++ {
			sec_prefix := cnv.Itoa(s)
			if pref, ok := entry[sec_prefix]; ok {
				sec_e := pref.(bson.M)
				xx, x_ok := sec_e[x].(time.Time)
				yy, y_ok := sec_e[y]
				if x_ok && y_ok {
					ts_ms := xx.UnixNano() / 1000000
					point := XY{ts_ms, yy}
					labelStr := c
					if len(labelStr) > 0 {
						if val, ok := res[labelStr]; ok {
							n := len(val)
							if cap(val) == n {
								res[labelStr] = make(Points, n, 2*n+1)
								copy(res[labelStr], val)
							}
							res[labelStr] = res[labelStr][0 : n+1]
							res[labelStr][n] = point

						} else {
							res[labelStr] = make(Points, 1, 200)
							res[labelStr][0] = point
						}
					}

				}
			}
		}
	}

	if err := cur.Close(); err != nil {
		return err, Empty_Graph
	}

	return nil, res
}

func GetKV(db, c, k string) string {
	s := masterSession.ms.Copy()
	coll := s.DB(db).C(c)
	defer s.Close()
	res := bson.M{}
	err := coll.Find(bson.M{"_id": k}).Select(bson.M{"_id": 1}).One(&res)
	if err != nil {
		log.Printf("Error retrieving value by key [%v].\n", k)
		log.Println(err.Error())
		return "N/A"
	}
	if str, ok := res["_id"].(string); ok {
		return str
	} else {
		return "N/A"
	}
}

func GetFilteredCollections(dbName, reStr string) (error, []string) {
	re, err := regexp.Compile(reStr)
	if err != nil {
		return err, []string{}
	}
	s := masterSession.ms.Copy()
	db := s.DB(dbName)
	defer s.Close()
	collections, err := db.CollectionNames()
	if err != nil {
		return err, collections
	}

	c_size := len(collections)

	if c_size > 0 {

		cnt := 0
		res := make([]string, c_size)
		for a := range collections {
			if re.MatchString(collections[a]) {
				res[cnt] = collections[a]
				cnt = cnt + 1
			}
		}

		return nil, res[:cnt]
	} else {
		return nil, []string{}
	}
}

//TODO will need to round up for upper bound
func objId(t time.Time, f round_time) string {
	nt := f(t)
	nt_sec := nt.Unix()
	dt_hex := cnv.FormatInt(nt_sec, 16)
	return dt_hex + zero_host
}

func t_rup_min(t time.Time) time.Time {
    return t.Add(time.Minute).Truncate(time.Minute)
}

func t_rdown_min(t time.Time) time.Time {
    return t.Truncate(time.Minute)
}

func (p Points) Len() int {
	return len(p)
}

func (p Points) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Points) Less(i, j int) bool {
	return p[i][0].(int) < p[j][0].(int)
}
