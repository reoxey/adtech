package mysql

import (
	"strings"

	my "github.com/pubnative/mysqldriver-go"
)

type Env struct {
	DB *my.DB
	obj
}

type obj struct {
	conn  *my.Conn
	list  []string
	assoc map[string]string
	table string
}

type Rows struct {
	*my.Rows
}

type Connector interface {
	Open() (Handler, error)
	Close() error
}

var _ Connector = (*Env) (nil)

type Handler interface {
	Select(...string) Handler
	Set(map[string]string) Handler
	Table(string) Handler
	Join(string) Handler
	Run(string) (Rows, error)
	Put() error
	Update(string) error
	Empty() Rows
}

var _ Handler = (*obj) (nil)

func Dial(dsn string, pool int) Connector {
	return Env {DB: my.NewDB(dsn, pool, -1)}
}

func (en Env) Open() (Handler, error) {
	c, e := en.DB.GetConn()
	if e != nil {
		return nil, e
	}
	return &obj{conn: c}, nil
}

func (o *obj) Select(f ...string) Handler {
	o.list = f
	return o
}

func (o *obj) Set(m map[string]string) Handler {
	o.assoc = m
	return o
}

func (o *obj) Table(t string) Handler {
	o.table = t
	return o
}

func (o *obj) Join(t string) Handler {
	o.table += " JOIN " + t
	return o
}

func (o *obj) Run(w string) (Rows, error) {
	if w == "" {
		w = "1"
	}
	r, e := o.conn.Query("SELECT "+strings.Join(o.list, ",")+" FROM "+o.table+" WHERE "+w)
	if e != nil {
		return Rows{}, e
	}
	return Rows{r}, nil
}

func (o *obj) Put() error {

	var val []string
	for k, v := range o.assoc {
		val = append(val, k+"='"+v+"'")
	}

	_, e := o.conn.Exec("INSERT "+o.table+" SET "+strings.Join(val, ","))

	return e
}

func (o *obj) Update(w string) error {

	var val []string
	for k, v := range o.assoc {
		val = append(val, k+"='"+v+"'")
	}

	if w == "" {
		w = "1"
	}

	_, e := o.conn.Exec("UPDATE "+o.table+" SET "+strings.Join(val, ",")+" WHERE "+w)

	return e
}

func (en Env) Close() error {
	return en.DB.PutConn(en.conn)
}

func (o *obj) Empty() Rows {
	return Rows{}
}
