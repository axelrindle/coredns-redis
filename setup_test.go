package redis

import (
	"context"
	"os"
	"testing"

	"github.com/coredns/caddy"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

    driver "github.com/gomodule/redigo/redis"
)

var container testcontainers.Container
var client driver.Conn

func TestMain(m *testing.M) {
	var err error

	// boot redis
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379:6379"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

    client, err = driver.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}

	// execute tests
	code := m.Run()

	// terminate redis
	err = container.Terminate(ctx)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func TestSetup(t *testing.T) {
	block := `
    redis {
        address localhost:6379
    }
    `

	c := caddy.NewTestController("dns", block)
	if err := setup(c); err != nil {
		t.Fatalf("Expected no errors, but got: %v", err)
	}
}

func TestSetupWrongAddress(t *testing.T) {
	block := `
    redis {
        address unknown:1337
    }
    `

	c := caddy.NewTestController("dns", block)
	if err := setup(c); err == nil {
		t.Fatalf("Expected errors, but got: %v", err)
	}
}
