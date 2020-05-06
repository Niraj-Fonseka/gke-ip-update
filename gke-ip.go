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
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
)

var (
	CredentialPath *string
	ProjectID      *string
	ClusterZone    *string
	ClusterID      *string
)

func main() {
	HandleArgs()

}

func SetCreds() {
	log.Println("Setting Google Credentials ")
	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/home/hungryotter/go/src/gke-ip-update/sa.json"); err != nil {
		log.Fatal(err)
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

	existingBlocks, err := GetExistingCidrBlock(projectID, clusterZone, clusterID, c, containerService)

	if err != nil {
		log.Fatal(err)
	}
	//https://godoc.org/google.golang.org/api/container/v1#ClusterUpdate
	//https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.zones.clusters/update
	//https://cloud.google.com/kubernetes-engine/docs/how-to/authorized-networks#api_2

	var updatedCidirBlocks []*container.CidrBlock
	cidrBlock := container.CidrBlock{
		CidrBlock:   "123.123.123.123/32",
		DisplayName: "TestHome",
	}

	for _, c := range existingBlocks {
		if c.DisplayName != cidrBlock.DisplayName {
			updatedCidirBlocks = append(updatedCidirBlocks, c)
		}
	}

	updatedCidirBlocks = append(updatedCidirBlocks, &cidrBlock)

	mAuthNetworkConfig := &container.MasterAuthorizedNetworksConfig{
		CidrBlocks: updatedCidirBlocks,
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

	resp, err := containerService.Projects.Zones.Clusters.Update(projectID, clusterZone, clusterID, rb).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Change code below to process the `resp` object:
	fmt.Printf("%#v\n", resp)
}
func HandleArgs() {
	CredentialPath = flag.String("path", "", "path for the google application credentials")
	ProjectID = flag.String("project", "", "project id")
	ClusterID = flag.String("string", "", "cluster id")
	ClusterZone = flag.String("zone", "", "zone where the master lives")
	flag.Parse()

	if *CredentialPath == "" {
		log.Fatal("No path for the service account provided")
	}

	if *ProjectID == "" {
		log.Fatal(("No project provided"))
	}

	if *ClusterZone == "" {
		log.Fatal("No zone provided")
	}

	if *ClusterID == "" {
		log.Fatal("ClusterID is not provided ")
	}
	fmt.Printf("Credentials Path : %s , ProjectID : %s , ClusterZone : %s , ClusteID : %s \n", *CredentialPath, *ProjectID, *ClusterZone, *ClusterID)
}

//https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.zones.clusters/get?apix_params=%7B%22projectId%22%3A%22agile-terra-275621%22%2C%22zone%22%3A%22us-central1-c%22%2C%22clusterId%22%3A%22projects-cluster%22%7D
func GetExistingCidrBlock(projectID string, zone string, clusterID string, client *http.Client, containerService *container.Service) ([]*container.CidrBlock, error) {
	ctx := context.Background()
	resp, err := containerService.Projects.Zones.Clusters.Get(projectID, zone, clusterID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return resp.MasterAuthorizedNetworksConfig.CidrBlocks, err

}
