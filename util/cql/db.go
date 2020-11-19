package cql

import (
	"log"
	"strings"
	"time"

	slog "github.com/RackSec/srslog"
	"github.com/fatih/structs"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"

	"corona/util/logger"
)

type CQL struct {
	Session *gocql.Session
	Elog    *slog.Writer
}

func (c *CQL) W(t, f string) qb.Cmp {
	switch t {
	case "eq":
		return qb.Eq(f)
	default:
		return qb.Cmp{}
	}
}

func Init(dsn map[string]string, hosts []string, el *slog.Writer) *CQL {
	cluster := gocql.NewCluster(hosts...)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: dsn["username"],
		Password: dsn["password"],
	}
	cluster.Keyspace = dsn["keyspace"]
	cluster.Consistency = gocql.One
	cluster.Timeout = 60 * time.Second
	var (
		session *gocql.Session
		e       error
	)
	if session, e = cluster.CreateSession(); e != nil {
		log.Println("db:Init", e)
	}
	return &CQL{session, el}
}

func (c *CQL) UUID() string {
	return gocql.TimeUUID().String()
}

func (c *CQL) Put(table string, s interface{}) {
	col := lower(structs.Names(s))
	stmt, names := qb.Insert(table).Columns(col...).ToCql()
	q := gocqlx.Query(c.Session.Query(stmt), names).BindStruct(s)

	if err := q.ExecRelease(); err != nil {
		logger.Err(c.Elog, err, "db:Put")
		log.Println("db:Put", err)
	}
}

func (c *CQL) PullOne(table string, s interface{}, m map[string]interface{}, w ...qb.Cmp) error {
	stmt, names := qb.Select(table).Where(w...).ToCql()
	q := gocqlx.Query(c.Session.Query(stmt), names).BindMap(m)
	if err := q.GetRelease(s); err != nil {
		if err.Error() != "not found" {
			logger.Err(c.Elog, err, "db:PullOne")
			log.Println("db:PullOne", err)
		}
		return err
	}
	return nil
}

func (c *CQL) PullAll(table string, s interface{}, m map[string]interface{}, w ...qb.Cmp) {
	stmt, names := qb.Select(table).Where(w...).ToCql()
	q := gocqlx.Query(c.Session.Query(stmt), names).BindMap(m)
	if err := q.SelectRelease(s); err != nil {
		logger.Err(c.Elog, err, "db:PullAll")
		log.Println("db:PullAll", err)
	}
}

func lower(s []string) []string {
	r := make([]string, 0)
	for _, a := range s {
		a = strings.ToLower(a)
		r = append(r, a)
	}
	return r
}
