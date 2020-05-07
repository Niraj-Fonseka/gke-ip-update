package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
)

var (
	CredentialPath     *string
	ProjectID          *string
	ClusterZone        *string
	ClusterID          *string
	Client             *http.Client
	NetworkDisplayName *string
	logFile            *os.File
)

func main() {
	initializeLogs()
	defer logFile.Close()
	Client = &http.Client{}
	handleArgs()
	initializeLocalStorage()
	ip, err := fetchIP()
	if err != nil {
		log.Fatal(err)
	}
	saveIP(ip)

	setCreds(*CredentialPath)
	setGKEIP(ip, *NetworkDisplayName)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go process(wg)

	wg.Wait()
}

func initializeLogs() {
	f, err := os.OpenFile(os.Getenv("HOME")+"/.gke_ip_update/gke_ip_update.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to initialize the log file")
	}

	logFile = f
}

func writeLog(message string) {
	if _, err := logFile.Write([]byte(message)); err != nil {
		log.Fatal("Unable to write to a log file")
	}
}

func process(wg *sync.WaitGroup) {

	for {
		ip, err := fetchIP()

		if err != nil {
			log.Println(err)
			break
		}
		savedIP := getIP()
		if savedIP != ip {
			saveIP(ip)
			log.Println("Ip change detected..")
			setGKEIP(ip, *NetworkDisplayName)
		}
		time.Sleep(10 * time.Second)
	}
	wg.Done()
}

func initializeLocalStorage() {
	homePath := os.Getenv("HOME")
	if homePath == "" {
		log.Fatal("Unable to get the path for HOME")
	}

	if _, err := os.Stat(homePath + "/.gke_ip_update"); os.IsNotExist(err) {
		err := os.Mkdir(homePath+"/.gke_ip_update", 0755)
		if err != nil {
			log.Fatal("Unable to create ,.gke_ip_update")
		}
	}

}

func saveIP(ip string) {
	err := ioutil.WriteFile(os.Getenv("HOME")+"/.gke_ip_update/ip.txt", []byte(ip), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getIP() string {
	ip, err := ioutil.ReadFile(os.Getenv("HOME") + "/.gke_ip_update/ip.txt")

	if err == os.ErrNotExist {
		log.Fatal(err)
	}

	cleanedIP := strings.TrimSuffix(string(ip), "\n")
	return cleanedIP
}

func fetchIP() (string, error) {
	resp, err := Client.Get("http://checkip.amazonaws.com/")

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	ip, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	cleanedIP := strings.TrimSuffix(string(ip), "\n")

	return cleanedIP, nil
}

func setCreds(path string) {
	log.Println("Setting Google Credentials ")
	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/home/hungryotter/go/src/gke-ip-update/sa.json"); err != nil {
		log.Fatal(err)
	}
}

func setGKEIP(ip, displayName string) {
	ctx := context.Background()

	c, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	containerService, err := container.New(c)
	if err != nil {
		log.Fatal(err)
	}

	existingBlocks, err := getExistingCidrBlock(*ProjectID, *ClusterZone, *ClusterID, c, containerService)

	if err != nil {
		log.Fatal(err)
	}

	var updatedCidirBlocks []*container.CidrBlock
	cidrBlock := container.CidrBlock{
		CidrBlock:   fmt.Sprintf("%s/32", ip),
		DisplayName: displayName,
	}

	for _, c := range existingBlocks {
		if c.DisplayName != cidrBlock.DisplayName {
			updatedCidirBlocks = append(updatedCidirBlocks, c)
		}
		if c.CidrBlock == fmt.Sprintf("%s/32", ip) {
			return
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
	}

	_, err = containerService.Projects.Zones.Clusters.Update(*ProjectID, *ClusterZone, *ClusterID, rb).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("IP successfully updated in the gke cluster")
}

func handleArgs() {
	CredentialPath = flag.String("path", "", "path for the google application credentials")
	ProjectID = flag.String("project", "", "project id")
	ClusterID = flag.String("cluster", "", "clusterid")
	ClusterZone = flag.String("zone", "", "zone where the master lives")
	NetworkDisplayName = flag.String("network_name", "", "DisplayName for the master authroized network")
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

	if *NetworkDisplayName == "" {
		log.Fatal("DisplayName is not provided")
	}

	fmt.Printf("Credentials Path : %s , ProjectID : %s , ClusterZone : %s , ClusteID : %s \n", *CredentialPath, *ProjectID, *ClusterZone, *ClusterID)
}

//https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.zones.clusters/get?apix_params=%7B%22projectId%22%3A%22agile-terra-275621%22%2C%22zone%22%3A%22us-central1-c%22%2C%22clusterId%22%3A%22projects-cluster%22%7D
func getExistingCidrBlock(projectID string, zone string, clusterID string, client *http.Client, containerService *container.Service) ([]*container.CidrBlock, error) {
	ctx := context.Background()
	resp, err := containerService.Projects.Zones.Clusters.Get(projectID, zone, clusterID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return resp.MasterAuthorizedNetworksConfig.CidrBlocks, err

}
