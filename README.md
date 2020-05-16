## gke-ip-update

### Description 
In GKE private clusters you have the option to authorize certian networks to access the master node so users can interact with the kube api server from a different network. If you are needing access to a kuberantes cluster from a network that does not have a static IP address ( your home ) you will have to keep changing your public IP address under Master Authorized Networks in the cluster everytime it changes. By running this app as a background job after giving the proper information it needs, it will continuously check and detect if your public IP address changes and update the Kubernates cluster automatically. 

### Build 
```
make build
```

the binary will be stored under gke-ip-update/bin directory 

### Stop 
```
make stop 
``` 

### Run ( as a background process )
```
./gke-ip-update --service-account "absolute path for the service account" --project "gcp-project-id" --zone "cluster-master-zone"  --cluster "cluster-name" --network_name "DisplayName for the network" & 
```

### Debugging 

When you run the application for the first time it will initialize a directory called .gke-ip-update at your $HOME. You can find your current ip address in `ip.txt` file and any logs related to the application will be stored in `gke_ip_update.log`. 