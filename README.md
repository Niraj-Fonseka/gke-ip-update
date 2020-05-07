## gke-ip-update

### Description 
In GKE private clusters you have the option to authorize certian networks to access the master node so users can interact with the kube api server from a different network. If you are needing access to a kuberantes cluster from a network that does not have a static IP address you will have to keep changing your public IP address under Master Authorized Networks in the cluster everytime it changes. By running this utility as a background job after giving the proper information it need, it will continiously check and detect if your public IP address changes, if it does it will update the Kubernates cluster automatically. 

### Build 
```
make build
```

### Run 
```
./gke-ip-update --path "absolute path for the service account" --project "gcp-project-id" --zone "cluster-master-zone"  --cluster "cluster-name" --network_name "DisplayName for the network" & 
```