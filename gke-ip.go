package main

// BEFORE RUNNING:
// ---------------
// 1. If not already done, enable the Kubernetes Engine API
//    and check the quota for your project at
//    https://console.developers.google.com/apis/api/container
// 2. This sample uses Application Default Credentials for authentication.
//    If not already done, install the gcloud CLI from
//    https://cloud.google.com/sdk/ and run
//    `gcloud beta auth application-default login`.
//    For more information, see
//    https://developers.google.com/identity/protocols/application-default-credentials
// 3. Install and update the Go dependencies by running `go get -u` in the
//    project directory.

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
)

func main() {
	ctx := context.Background()

	c, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	containerService, err := container.New(c)
	if err != nil {
		log.Fatal(err)
	}

	// Deprecated. The Google Developers Console [project ID or project
	// number](https://support.google.com/cloud/answer/6158840).
	// This field has been deprecated and replaced by the name field.
	projectId := "my-project-id" // TODO: Update placeholder value.

	// Deprecated. The name of the Google Compute Engine
	// [zone](/compute/docs/zones#available) in which the cluster
	// resides.
	// This field has been deprecated and replaced by the name field.
	zone := "my-zone" // TODO: Update placeholder value.

	// Deprecated. The name of the cluster to upgrade.
	// This field has been deprecated and replaced by the name field.
	clusterId := "my-cluster-id" // TODO: Update placeholder value.

	rb := &container.UpdateClusterRequest{
		// TODO: Add desired fields of the request body. All existing fields
		// will be replaced.
	}

	resp, err := containerService.Projects.Zones.Clusters.Update(projectId, zone, clusterId, rb).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Change code below to process the `resp` object:
	fmt.Printf("%#v\n", resp)
}
