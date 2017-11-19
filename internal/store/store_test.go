package store

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/izumin5210/ro/types"
	dockertest "gopkg.in/ory-am/dockertest.v3"
)

// Types
// ================================================================

type TestPost struct {
	ID    uint64 `redis:"id"`
	Title string `redis:"title"`
	Body  string `redis:"body"`
}

func (p *TestPost) GetKeyPrefix() string {
	return ""
}

func (p *TestPost) GetKeySuffix() string {
	return fmt.Sprint(p.ID)
}

// Test funcs
// ================================================================

func TestSet(t *testing.T) {
	post := &TestPost{
		ID:    1,
		Title: "post 1",
		Body:  "This is a post 1.",
	}

	cnf := &types.StoreConfig{}
	store, err := New(redisPool.Get, &TestPost{}, cnf)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	err = store.Set(post)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	conn := redisPool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", "*"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if got, want := len(keys), 1; err != nil {
		t.Errorf("Stored keys was %d, want %d", got, want)
	}

	v, err := redis.Values(conn.Do("HGETALL", "TestPost:1"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	gotPost := &TestPost{}
	err = redis.ScanStruct(v, gotPost)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if got, want := gotPost, post; !reflect.DeepEqual(got, want) {
		t.Errorf("Stored post is %v, want %v", got, want)
	}
}

func TestGet(t *testing.T) {
	store, err := New(redisPool.Get, &TestPost{}, &types.StoreConfig{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	post := &TestPost{
		ID:    1,
		Title: "post 1",
		Body:  "This is a post 1.",
	}
	key := "TestPost:1"

	conn := redisPool.Get()
	defer conn.Close()
	conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(post)...)

	gotPost := &TestPost{ID: 1}
	err = store.Get(gotPost)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if got, want := gotPost, post; !reflect.DeepEqual(got, want) {
		t.Errorf("Stored post is %v, want %v", got, want)
	}
}

// Setup and Teardown
// ================================================================

var redisPool *redis.Pool

func TestMain(m *testing.M) {
	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("redis", "4.0.2-alpine", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(fmt.Sprintf("redis://localhost:%s", resource.GetPort("6379/tcp")))
		},
	}

	if err = pool.Retry(func() error {
		conn := redisPool.Get()
		defer conn.Close()
		_, err := conn.Do("PING")

		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	exitCode := m.Run()

	err = redisPool.Close()
	if err != nil {
		log.Fatalf("Failed to close redis pool: %s", err)
	}
	err = pool.Purge(resource)
	if err != nil {
		log.Fatalf("Failed to purge docker pool: %s", err)
	}

	os.Exit(exitCode)
}
