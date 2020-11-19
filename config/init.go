package config

import (
	"io/ioutil"
	"os"

	slog "github.com/RackSec/srslog"
	"gopkg.in/yaml.v2"

	"corona/util/cql"
	"corona/util/mysql"
	"corona/util/redis"
	"corona/util/throttler"
	"corona/util/userdetect"
)

// Obj config details
type Obj struct {
	LogE    *slog.Writer
	LogV    *slog.Writer
	DB      mysql.Connector
	QL      *cql.CQL
	RD      *redis.Conn
	Limiter *throttler.IPRateLimiter
	QPS     map[string]int
}

type config struct {
	Log struct {
		Error   string `yaml:"error"`
		Verbose string `yaml:"verbose"`
		Host    string `yaml:"server"`
	}
	Dbs struct {
		Ip2l string `yaml:"ip2l"`
		D51  string `yaml:"d51"`
	}
	Redis struct {
		Addr string `yaml:"addr"`
	}
	Mysql struct {
		Dsn string `yaml:"dsn"`
	}
	Cassandra struct {
		KeySpace string `yaml:"keyspace"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}
	Endpoints struct {
		Cassandra []string `yaml:"cassandra"`
	}
}

func read() config {
	one := config{}

	f, e := os.Open("./config/local.yml")
	if e != nil {
		f, _ = os.Open("../config/local.yml")
	}
	b, _ := ioutil.ReadAll(f)

	e = yaml.Unmarshal(b, &one)
	if e != nil {
		panic(e)
	}

	return one
}

// Init initialise Obj with database connections, logs, rate limiter etc.
func Init() Obj {
	c := read()

	userdetect.Init(c.Dbs.D51, c.Dbs.Ip2l)

	dsn := map[string]string{
		"username": c.Cassandra.Username,
		"password": c.Cassandra.Password,
		"keyspace": c.Cassandra.KeySpace,
	}

	le, _ := slog.Dial("udp", c.Log.Host+":514", slog.LOG_ERR, c.Log.Error)
	lv, _ := slog.Dial("udp", c.Log.Host+":514", slog.LOG_INFO, c.Log.Verbose)

	return Obj{
		LogE:    le,
		LogV:    lv,
		DB:      mysql.Dial(c.Mysql.Dsn, 1000),
		QL:      cql.Init(dsn, c.Endpoints.Cassandra, le),
		RD:      redis.Connect(c.Redis.Addr),
		Limiter: throttler.NewIPRateLimiter(),
		QPS:     nil,
	}
}
