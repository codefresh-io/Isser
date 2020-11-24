# VENONACTL
Codefresh installer of Kubernetes YAML onto the cluster

## Version 1.x.x
Version 1.x.x is is about to be released soon, read more about migration from older version [here](#Migration)  
Meanwhile 1.x.x is to release and makred as pre-release we will maintain 2 branches:
* `master` - the previous version ( `version < 1.0.0` )
  * we will keep maintaing if (bugs, security issues) - this version will be intalled when installing `venona` on MacOS using brew
  * `quay.io/codefresh/venona:latest` will refer to this branch
* `release-1.0` it the new release, which will be used when running Codefresh CLI to install the agent
We highly suggest to use [Codefresh official CLI](https://codefresh-io.github.io/cli/) to install the agent:
```bash
codefresh runner init
```

The last command will:  
1. Install the agent on the namespace `codefresh` (as you choose)
2. Install the runtime on the same namespace
3. Attach the runtime to the agent
4. Register cluster on codefresh platform
5. Create and run demo pipeline

It is still possible, for advanced users to install all manually, for example:
One process of Venona can manage multiple runtime environments
NOTE: Please make sure that the process where Venona is installed there is a network connection to the clusters where the runtimes will be installed
```bash
# 1. Create namespace for the agent: 
kubectl create namespace codefresh-agent

# 2. Install the agent on the namespace ( give your agent a unique):
# Print a token that the Venona process will be using.
codefresh create agent $NAME
codefresh install agent --token $TOKEN --kube-namespace codefresh-agent

# 3. Create namespace for the first runtime:
kubectl create namespace codefresh-runtime-1

# 4. Install the first runtime on the namespace
# 5. the runtime name is printed
codefresh install runtime --kube-namespace codefresh-runtime-1

# 6. Attach the first runtime to agent:
codefresh attach runtime --agent-name $AGENT_NAME --agent-kube-namespace codefresh-agent --runtime-name $RUNTIME_NAME --kube-namespace codefresh-runtime-1

# 7. Restart the venona pod in namespace `codefresh-agent`
kubectl delete pods $VENONA_POD

# 8. Create namespace for the second runtime
kubectl create namespace codefresh-runtime-2

# 9. Install the second runtime on the namespace
codefresh install runtime --kube-namespace codefresh-runtime-2

# 10. Attach the second runtime to agent and restart the Venoa pod automatically
codefresh attach runtime --agent-name $AGENT_NAME --agent-kube-namespace codefresh-agent --runtime-name $RUNTIME_NAME --runtime-kube-namespace codefresh-runtime-1 --restart-agent

```

## Migration
Migrating from Venona `< 1.x.x` to `> 1.x.x` is not done automatically, please run the follwing
```bash
codefresh runner upgrade
```


## Installation

### Prerequisite:
* [Kubernetes](https://kubernetes.io/docs/tasks/tools/install-kubectl/) - Used to create resource in your K8S cluster
  * Kube Version > 1.10:
    * [Instuction](#Install-on-cluster-version-<-1.10) to install on cluster version < 1.10
  * Disk size 50GB per node
* [Codefresh](https://codefresh-io.github.io/cli/) - Used to create resource in Codefresh
  * Authenticated context exist under `$HOME/.cfconfig` or authenticate with [Codefesh CLI](https://codefresh-io.github.io/cli/getting-started/#authenticate)

### Install venona

* Download [venona's](https://github.com/codefresh-io/venona/releases) binary
  * With homebrew:
    * `brew tap codefresh-io/venona`
    * `brew install venona`


#### Install on cluster version < 1.10
* Make sure the `PersistentLocalVolumes` [feature gate](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/) is turned on
* Venona's agent is trying to load avaliables apis using api `/openapi/v2` endpoint
Add this endpoint to ClusterRole `system:discovery` under `rules[0].nonResourceURLs`
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:discovery
rules:
- nonResourceURLs:
  - ...other_resources
  - /openapi
  - /openapi/*
  verbs:
  - get
```
#### Install using --values <values.yaml>
`codefresh runner init --values <values.yaml> [parameters]`
the values from values.yaml are applied to the templates in [pkg/templates/kubernetes](pkg/templates/kubernetes)

See BuildValues() func in [store.go](pkg/store/store.go) for the format  
Example with explaination is in [values-example.yaml](values-example.yaml) 

#### Install on GCP
  * Make sure your user has `Kubernetes Engine Cluster Admin` role in google console
  * Bind your user with cluster-admin kubernetes clusterrole
    > `kubectl create clusterrolebinding NAME --clusterrole cluster-admin --user YOUR_USER`

#### Pipeline Storage with docker cache support

##### **GKE LocalSSD**
*Prerequisite:* [GKE custer with local SSD](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/local-ssd)

*Install venona for using GKE Local SSD:*
```
codefresh install runtime [options] \
                    --set-value=Storage.LocalVolumeParentDir=/mnt/disks/ssd0/codefresh-volumes 
                    --kube-node-selector=cloud.google.com/gke-local-ssd=true
```

##### **GCE Disks** 
*Prerequisite:* volume provisioner (dind-volume-provisioner) should have permissions to create/delete/get of google disks
There are 3 options to provide cloud credentials on GCE:
* run venona dind-volume-provisioniner on node with iam role which is allowed to create/delete/get of google disks
* create Google Service Account with ComputeEngine.StorageAdmin, download its key and pass it to venona installed with `--set-file=Storage.GooogleServiceAccount=/path/to/google-service-account.json`
* use [Google Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity) to assign iam role to `volume-provisioner-venona` service account 

*Note*: Builds will be running in single availability zone, so you must to specify AvailabilityZone params


*Install venona for using GKE Disks:*
```
codefresh install runtime [options] \
                    --set-value=Storage.Backend=gcedisk \
                    --set-value=Storage.AvailabilityZone=us-central1-a \
                    --kube-node-selector=failure-domain.beta.kubernetes.io/zone=us-central1-a \
                    [--set-file=Storage.GoogleServiceAccount=/path/to/google-service-account.json]
```

##### **Amazon EBS**

*Prerequisite:* volume provisioner (dind-volume-provisioner) should have permissions to create/delete/get of aws ebs
Minimal iam policy for dind-volume-provisioner:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:AttachVolume",
        "ec2:CreateSnapshot",
        "ec2:CreateTags",
        "ec2:CreateVolume",
        "ec2:DeleteSnapshot",
        "ec2:DeleteTags",
        "ec2:DeleteVolume",
        "ec2:DescribeInstances",
        "ec2:DescribeSnapshots",
        "ec2:DescribeTags",
        "ec2:DescribeVolumes",
        "ec2:DetachVolume"
      ],
      "Resource": "*"
    }
  ]
}
```

There are 3 options to provide cloud credentials on AWS:
* run venona dind-volume-provisioniner on node with the iam role - use `--set-value Storage.VolumeProvisioner.NodeSelector=node-label=value` option
* create AWS IAM User, assign it the permissions above and suppy aws credentials to venona installer `--set-value=Storage.AwsAccessKeyId=ABCDF --set-value=Storage.AwsSecretAccessKey=ZYXWV`

* use [Aws Identity for Service Account](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html) to assign iam role to `volume-provisioner-venona` service account

*Notes*: 
- Builds will be running in single availability zone, so you must specify AvailabilityZone parameter `--set-value=Storage.AvailabilityZone=<aws-az>` and build-node-selector `--build-node-selector=failure-domain.beta.kubernetes.io/zone=<aws-az>` in case of multizone cluster

- We support both [in-tree ebs](https://kubernetes.io/docs/concepts/storage/volumes/#awselasticblockstore) (`--set-value=Storage.Backend=ebs`) volumes and ebs-csi(https://github.com/kubernetes-sigs/aws-ebs-csi-driver) (`--set-value=Storage.Backend=ebs-csi`)

*Install Command to run pipelines on ebs volumes*
```
codefresh install runtime [options] \
                    --set-value=Storage.Backend=ebs \
                    --set-value=Storage.AvailabilityZone=us-east-1d \
                    --kube-node-selector=failure-domain.beta.kubernetes.io/zone=us-east-1d \
                    [--set-value Storage.VolumeProvisioner.NodeSelector=kubernetes.io/role=master] \
                    [--set-value Storage.AwsAccessKeyId=ABCDF --set-value Storage.AwsSecretAccessKey=ZYXWV]\
                    [--set-value=Storage.Encrypted=true] \
                    [--set-value=Storage.KmsKeyId=<key id>]
```

##### **Azure Disk**
*Prerequisite:* volume provisioner (dind-volume-provisioner) should have permissions to create/delete/get Auzure Disks

Minimal iam Role for dind-volume-provisioner - dind-volume-provisioner-role.json:
```json
{
    "Name": "CodefreshDindVolumeProvisioner",
    "Description": "Perform create/delete/get disks",
    "IsCustom": true,
    "Actions": [
        "Microsoft.Compute/disks/read",
        "Microsoft.Compute/disks/write",
        "Microsoft.Compute/disks/delete"

    ],
    "AssignableScopes": ["/subscriptions/<your-subsripton_id>"]
}
```
If you use AKS with managed [identities for node group](https://docs.microsoft.com/en-us/azure/aks/use-managed-identity), you can run the script below to assign CodefreshDindVolumeProvisioner role to aks node identity: 
```bash
export ROLE_DEFINITIN_FILE=dind-volume-provisioner-role.json
export SUBSCRIPTION_ID=$(az account show --query "id" | xargs echo )
export RESOURCE_GROUP=codefresh-rt1
export AKS_NAME=codefresh-rt1
export LOCATION=$(az aks show -g $RESOURCE_GROUP -n $AKS_NAME --query location | xargs echo)
export NODES_RESOURCE_GROUP=MC_${RESOURCE_GROUP}_${AKS_NAME}_${LOCATION}
export NODE_SERVICE_PRINCIPAL=$(az aks show -g $RESOURCE_GROUP -n $AKS_NAME --query identityProfile.kubeletidentity.objectId | xargs echo)

az role definition create --role-definition @${ROLE_DEFINITIN_FILE}
az role assignment create --assignee $NODE_SERVICE_PRINCIPAL --scope /subscriptions/$SUBSCRIPTION_ID/resourceGroups/$NODES_RESOURCE_GROUP --role CodefreshDindVolumeProvisioner
```

Now create runner with `--set-value Storage.Backend=azuredisk --set Storage.VolumeProvisioner.MountAzureJson=true`:
```
codefresh runner init --set-value Storage.Backend=azuredisk --set Storage.VolumeProvisioner.MountAzureJson=true 
```
Or using runner-values.yaml file like below:  
```yaml
# CodefreshHost: https://g.codefresh.io
# Token: ******
# Namespace: default
# Context: codefresh-rt1 
# RuntimeInCluster: true
Storage:
  Backend: azuredisk
  VolumeProvisioner:
    MountAzureJson: true
```
```
codefresh runner init --values runner-values.yaml 
```

#### Kubernetes RBAC
Installation of Venona on Kubernetes cluster installing 2 groups of objects,
Each one has own RBAC needs and therefore, created roles(and cluster-roles)
The resource descriptors are avaliable [here](https://github.com/codefresh-io/venona/tree/master/venonactl/templates/kubernetes)
List of the resources that will be created
* Agent (grouped by `/.*.venona.yaml/`)
  * `service-account.re.yaml` - The service account that the Venona pod will use to create the resource on the runtime namespace(the resoucre installed on the runtime namespace)
  * `role.re.yaml` - Allow to GET, CREATE and DELETE pods and persistentvolumeclaims
  * `role-binding.re.yaml` - The agent is spinning up pods and pvc, this biniding binds `role.venona.yaml` to `service-account.venona.yaml`
  * `cluster-role-binding.venona.yaml` - The agent discovering K8S apis by calling to `openapi/v2`, this ClusterRoleBinding binds  bootstraped ClusterRole by Kubernetes `system:discovery` to `service-account.venona.yaml`. This role has only permissions to make a GET calls to non resources urls
* Runtime-environment (grouped by `/.*.re.yaml/`) Kubernetes controller that spins up all required resources to provide a good caching expirience during pipeline execution
  * `service-account.dind-volume-provisioner.re.yaml` - The service account that the controller will use
  * `cluster-role.dind-volume-provisioner.re.yaml` Defines all the permission needed for the controller to operate correctly
  * `cluster-role-binding.dind-volume-provisioner.yaml` - Binds the ClusterRole to `service-account.dind-volume-provisioner.re.yaml`
