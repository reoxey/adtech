package redis

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

type Conn struct {
	c *redis.Client
}

func Connect(addr string) *Conn {

	cl := redis.NewClient(&redis.Options{
		Addr:     addr + ":6379",
		Password: "",
		DB:       0,
	})

	pong, e := cl.Ping().Result()
	fmt.Println(pong, e)

	return &Conn{c: cl}
}

func (r *Conn) Set(k, v string) {
	go func() {
		r.c.Set(k, v, 0)
	}()
}

func (r *Conn) SetEx(k, v string, t time.Duration) {
	go func() {
		r.c.Set(k, v, t)
	}()
}

func (r *Conn) Get(k string) (string, error) {
	return r.c.Get(k).Result()
}

func (r *Conn) GetAr(k string) ([]string, error) {
	s, e := r.c.Get(k).Result()
	if e != nil {
		return []string{}, e
	}
	return strings.Split(s, ","), nil
}

func (r *Conn) GetJSON(k string, v interface{}) (e error) {
	s, e := r.c.Get(k).Result()
	if e != nil {
		return
	}
	return json.Unmarshal([]byte(s), v)
}
