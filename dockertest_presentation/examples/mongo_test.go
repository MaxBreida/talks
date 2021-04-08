package examples

import (
	"context"
	"fmt"
	"testing"

	"github.com/ory/dockertest/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Test_Mongo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database tests (short!)")
	}

	// START1 OMIT
	var db *mongo.Client
	var err error

	// Create the pool (docker instance).
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not connect to docker: %s", err)
	}

	// Start the container.
	container, err := pool.Run("mongo", "4.2", nil) // last param: optional env vars
	if err != nil {
		t.Fatalf("Could not start container: %s", err)
	}

	// Ensure container is cleaned up.
	t.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			t.Fatalf("failed to cleanup mongo container: %s", err)
		}
	})
	// END1 OMIT

	// START2 OMIT
	// Wait for the container to start - we'll retry connections in a loop
	// because the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		ctx := context.Background()

		db, err = mongo.Connect(ctx, &options.ClientOptions{
			Hosts: []string{fmt.Sprintf(
				"%s:%s",
				container.GetBoundIP("27017/tcp"),
				container.GetPort("27017/tcp"),
			)},
		})
		if err != nil {
			return err
		}

		// if ping succeeds, we can continue our actual test
		return db.Ping(ctx, nil)
	}); err != nil { // timeout after 1 min
		t.Fatalf("Could not connect to docker: %s", err)
	}
	// END2 OMIT

	// from here on out, you can use db to test your mongodb
}
