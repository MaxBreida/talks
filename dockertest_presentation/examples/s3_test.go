package examples

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Clarilab/s3-client"
	"github.com/minio/minio-go/v6"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func Test_S3(t *testing.T) {
	// START1 OMIT
	var minioClient *minio.Client
	var err error

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	// docker run -p 9000:9000 \
	//  -e "MINIO_ACCESS_KEY=MYACCESSKEY" \
	//  -e "MINIO_SECRET_KEY=MYSECRETKEY" \
	//  minio/minio:latest server /data
	options := &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"}, // directory where files are stored
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9000/tcp": {{HostPort: "9000"}},
		},
		Env: []string{"MINIO_ACCESS_KEY=MYACCESSKEY", "MINIO_SECRET_KEY=MYSECRETKEY"},
	}
	// END1 OMIT

	// START2 OMIT
	container, err := pool.RunWithOptions(options, func(config *docker.HostConfig) {
		config.AutoRemove = true // set AutoRemove so that stopped container removes itself
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no", // tell docker not to restart container after it stops
		}
	})
	if err != nil {
		t.Fatalf("could not start container: %s", err)
	}

	container.Expire(60) // Tell docker to hard kill the container in 60 seconds

	// Ensure container is cleaned up.
	t.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			t.Fatalf("failed to cleanup s3 container: %s", err)
		}
	})

	url := fmt.Sprintf(
		"%s:%s",
		container.GetBoundIP("9000/tcp"),
		container.GetPort("9000/tcp"),
	)
	// END2 OMIT

	// START3 OMIT
	// Wait for the container to start - we'll retry connections in a loop
	// the minio client does not do service discovery for you
	// (it does not check if connection can be established), so we have to use the health check
	if err = pool.Retry(func() error {
		url := fmt.Sprintf("http://%s/minio/health/live", url)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code not OK")
		}
		return nil
	}); err != nil {
		t.Fatalf("could not connect to docker: %s", err)
	}

	// check your errors pls
	minioClient, err = minio.New(url, "MYACCESSKEY", "MYSECRETKEY", false)
	err = minioClient.MakeBucket("my-documents", "")
	s3Client, err := s3.NewClient(url, "MYACCESSKEY", "MYSECRETKEY", "my-documents", false)
	// END3 OMIT

	// from here on out, you can use s3Client to test your implementation
	s3Client.DownloadFile("")
}