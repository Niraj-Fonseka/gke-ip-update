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
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
)

func main() {

	project := os.Getenv("GCP_PROJECT")

	if project == "" {
		log.Fatal("Project is not set")
	}
	clusterZone := os.Getenv("GCP_KUBE_MASTER_ZONE")

	if clusterZone == "" {
		log.Fatal("Cluster Zone is not set")
	}

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
	projectId := project // TODO: Update placeholder value.

	// Deprecated. The name of the Google Compute Engine
	// [zone](/compute/docs/zones#available) in which the cluster
	// resides.
	// This field has been deprecated and replaced by the name field.
	zone := "us-central1-c" // TODO: Update placeholder value.

	// Deprecated. The name of the cluster to upgrade.
	// This field has been deprecated and replaced by the name field.
	clusterId := "projects-cluster" // TODO: Update placeholder value.

	//https://godoc.org/google.golang.org/api/container/v1#ClusterUpdate
	//https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.zones.clusters/update
	//https://cloud.google.com/kubernetes-engine/docs/how-to/authorized-networks#api_2
	cidrBlock := container.CidrBlock{
		CidrBlock:   "47.221.172.101/32",
		DisplayName: "NirajHome",
	}
	mAuthNetworkConfig := &container.MasterAuthorizedNetworksConfig{
		CidrBlocks: []*container.CidrBlock{&cidrBlock},
		Enabled:    true,
	}
	clusterUpdate := container.ClusterUpdate{

		DesiredMasterAuthorizedNetworksConfig: mAuthNetworkConfig,
	}

	rb := &container.UpdateClusterRequest{
		Update: &clusterUpdate,
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

func GetExistingCidrBlock() {

}
