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
	credentialPath     *string
	projectID          *string
	clusterZone        *string
	clusterID          *string
	client             *http.Client
	networkDisplayName *string
	logFile            *os.File
)

func main() {
	defer logFile.Close()

	client = &http.Client{}
	handleArgs()
	initializeLocalStorage()
	initializeLogs()
	ip, err := findPublicIP()
	if err != nil {
		log.Fatal(err)
	}
	saveIP(ip)

	setCreds(*credentialPath)
	setGKEIP(ip, *networkDisplayName)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go run(wg)

	wg.Wait()
}

//initialize log file
func initializeLogs() {

	if _, err := os.Stat(os.Getenv("HOME") + "/.gke_ip_update/gke_ip_update.log"); os.IsNotExist(err) {
		if _, err := os.Create(os.Getenv("HOME") + "/.gke_ip_update/gke_ip_update.log"); err != nil {
			log.Fatal("Cant Create log file : ", err)
		}

	}
	f, err := os.OpenFile(os.Getenv("HOME")+"/.gke_ip_update/gke_ip_update.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Unable to initialize the log file : ", err)
	}

	logFile = f
}

//write log to file
func writeLog(message string) {
	if _, err := logFile.Write([]byte(message)); err != nil {
		log.Fatal("Unable to write to a log file")
	}
}

//runs a job that checks the ip every 3 minutes and updates the gke cluster if needed
func run(wg *sync.WaitGroup) {

	for {
		ip, err := findPublicIP()

		if err != nil {
			log.Println(err)
			break
		}
		savedIP := getIP()
		if savedIP != ip {
			writeLog(fmt.Sprintf("IP change detected from : %s , to : %s ", savedIP, ip))
			saveIP(ip)
			setGKEIP(ip, *networkDisplayName)
		}
		time.Sleep(3 * time.Minute)
	}
	wg.Done()
}

//create a directory for maintaing state / metadata
func initializeLocalStorage() {
	homePath := os.Getenv("HOME")
	if homePath == "" {
		log.Fatal("Unable to get the path for HOME")
	}

	if _, err := os.Stat(homePath + "/.gke_ip_update"); os.IsNotExist(err) {
		err := os.Mkdir(homePath+"/.gke_ip_update", 0755)
		if err != nil {
			log.Fatal("Unable to create .gke_ip_update directory")
		}
	}

}

//save the ip to the local state
func saveIP(ip string) {
	err := ioutil.WriteFile(os.Getenv("HOME")+"/.gke_ip_update/ip.txt", []byte(ip), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

//read ip from local state
func getIP() string {
	ip, err := ioutil.ReadFile(os.Getenv("HOME") + "/.gke_ip_update/ip.txt")

	if err == os.ErrNotExist {
		log.Fatal(err)
	}

	cleanedIP := strings.TrimSuffix(string(ip), "\n")
	return cleanedIP
}

//find the public IP address
func findPublicIP() (string, error) {
	resp, err := client.Get("http://checkip.amazonaws.com/")

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

//get GOOGLE_APPLICATION_CREDENTIALS using the path given by the user
func setCreds(path string) {
	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/home/hungryotter/go/src/gke-ip-update/sa.json"); err != nil {
		log.Fatal(err)
	}
}

//if the IP change has been detected update the list of Master Authroized Networks in the GKE cluster
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

	existingBlocks, err := getExistingCidrBlock(*projectID, *clusterZone, *clusterID, c, containerService)

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

	_, err = containerService.Projects.Zones.Clusters.Update(*projectID, *clusterZone, *clusterID, rb).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	writeLog("IP successfully updated in the gke cluster")
}

//Parsing arguments at the start of the app
func handleArgs() {
	credentialPath = flag.String("path", "", "path for the google application credentials")
	projectID = flag.String("project", "", "project id")
	clusterID = flag.String("cluster", "", "clusterid")
	clusterZone = flag.String("zone", "", "zone where the master lives")
	networkDisplayName = flag.String("network_name", "", "DisplayName for the master authroized network")
	flag.Parse()

	if *credentialPath == "" {
		log.Fatal("No path for the service account provided")
	}

	if *projectID == "" {
		log.Fatal(("No project provided"))
	}

	if *clusterZone == "" {
		log.Fatal("No zone provided")
	}

	if *clusterID == "" {
		log.Fatal("ClusterID is not provided ")
	}

	if *networkDisplayName == "" {
		log.Fatal("DisplayName is not provided")
	}

}

//https://cloud.google.com/kubernetes-engine/docs/reference/rest/v1/projects.zones.clusters/get?apix_params=%7B%22projectId%22%3A%22agile-terra-275621%22%2C%22zone%22%3A%22us-central1-c%22%2C%22clusterId%22%3A%22projects-cluster%22%7D
//fetch the existing networks in the GKE cluster
func getExistingCidrBlock(projectID string, zone string, clusterID string, client *http.Client, containerService *container.Service) ([]*container.CidrBlock, error) {
	ctx := context.Background()
	resp, err := containerService.Projects.Zones.Clusters.Get(projectID, zone, clusterID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return resp.MasterAuthorizedNetworksConfig.CidrBlocks, err

}
