package ro_test

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/alunir/ro"
	"github.com/alunir/ro/rq"
	rotesting "github.com/alunir/ro/testing"
)

// Setup and cleanup
// ----------------------------------------------------------------
const (
	POST_SERIALIZED_KEY = "POST_SERIALIZED"
)

func setup_serialized() {
	now = time.Now()
	postStore = ro.New(pool, &rotesting.Post_Serialized{}, ro.WithHashStore(false), ro.WithKeyPrefix(POST_SERIALIZED_KEY))

	postStore.Put(context.TODO(), []*rotesting.Post_Serialized{
		{
			ID:        1,
			UserID:    1,
			Title:     "post 1",
			Body:      "This is a post 1",
			CreatedAt: now.UnixNano(),
		},
		{
			ID:        2,
			UserID:    2,
			Title:     "post 2",
			Body:      "This is a post 2",
			CreatedAt: now.Add(-24 * 60 * 60 * time.Second).UnixNano(),
		},
		{
			ID:        3,
			UserID:    1,
			Title:     "post 3",
			Body:      "This is a post 3",
			CreatedAt: now.Add(24 * 60 * 60 * time.Second).UnixNano(),
		},
		{
			ID:        4,
			UserID:    1,
			Title:     "post 4",
			Body:      "This is a post 4",
			CreatedAt: now.Add(-24 * 60 * 60 * time.Second).UnixNano(),
		},
	}, 600)
}

// Examples
// ----------------------------------------------------------------

func Example_Store_Set_Serialized() {
	defer cleanup()
	postStore = ro.New(pool, &rotesting.Post_Serialized{}, ro.WithHashStore(false), ro.WithKeyPrefix(POST_SERIALIZED_KEY))

	postStore.Put(context.TODO(), []*rotesting.Post_Serialized{
		{
			ID:        1,
			UserID:    1,
			Title:     "post 1",
			Body:      "This is a post 1",
			CreatedAt: now.UnixNano(),
		},
		{
			ID:        2,
			UserID:    1,
			Title:     "post 2",
			Body:      "This is a post 2",
			CreatedAt: now.Add(-24 * 60 * 60 * time.Second).UnixNano(),
		},
	}, 600)

	conn := pool.Get()
	defer conn.Close()

	// tbl_keys, _ := redis.Strings(conn.Do("KEYS", "*"))
	// fmt.Printf("tbl_keys: %v\n", tbl_keys)
	// s3, _ := redis.Strings(conn.Do("HMGET", "POST_SERIALIZED", "2", "1"))
	// fmt.Println(s3)

	keys, _ := redis.Strings(conn.Do("ZRANGE", POST_SERIALIZED_KEY+"/user:1", 0, -1))
	args := []interface{}{POST_SERIALIZED_KEY}
	for _, k := range keys {
		args = append(args, k)
	}
	vv, _ := redis.Values(conn.Do("HMGET", args...))
	for _, v := range vv {
		p := &rotesting.Post_Serialized{}
		p.Deserialized(v.([]byte))
		fmt.Println(p.Body)
	}
	// Output:
	// This is a post 2
	// This is a post 1
}

func Example_Store_Get_Serialized() {
	setup_serialized()
	defer cleanup()

	post := &rotesting.Post_Serialized{ID: 1}
	postStore.Get(context.TODO(), post)
	fmt.Println(post.Body)
	// Output:
	// This is a post 1
}

func Example_Store_List_Serialized() {
	setup_serialized()
	defer cleanup()

	conn := pool.Get()
	defer conn.Close()
	// keys, _ := redis.Strings(conn.Do("KEYS", "*"))
	// fmt.Println(keys)
	// s2, _ := redis.Strings(conn.Do("ZREVRANGEBYSCORE", "POST_SERIALIZED/recent", "+inf", time.Now().UnixNano()))
	// fmt.Println(s2)
	// s, _ := redis.Strings(conn.Do("HGETALL", "POST_SERIALIZED"))
	// fmt.Println(s)
	// s3, _ := redis.Strings(conn.Do("HMGET", "POST_SERIALIZED", "3", "1"))
	// fmt.Println(s3)

	var posts []rotesting.Post_Serialized
	postStore.List(context.TODO(), &posts, rq.Key("recent"), rq.GtEq(now.UnixNano()), rq.Reverse())
	fmt.Println(posts[0].Body)
	fmt.Println(posts[1].Body)
	// Output:
	// This is a post 3
	// This is a post 1
}

func Example_Store_Count_Serialized() {
	setup_serialized()
	defer cleanup()

	cnt, _ := postStore.Count(context.TODO(), rq.Key("user", 1), rq.LtEq(now.UnixNano()))
	fmt.Println(cnt)
	// Output:
	// 2
}
