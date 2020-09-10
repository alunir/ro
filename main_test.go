package ro_test

import (
	"log"
	"os"
	"testing"

	rotesting "github.com/alunir/ro/testing"
)

var pool *rotesting.Pool

func TestMain(m *testing.M) {
	pool = rotesting.MustCreate()

	code := m.Run()

	pool.MustClose()

	os.Exit(code)
}

func teardown(t *testing.T) {
	if err := pool.Cleanup(); err != nil {
		log.Fatalf("Failed to flush redis: %s", err)
	}
}
