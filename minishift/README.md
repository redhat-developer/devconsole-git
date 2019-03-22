## Running Git Service on Minishift

These instructions will help you run the service using MiniShift.

### Prerequisites

[MiniShift v1.27.0](https://docs.okd.io/latest/minishift/getting-started/installing.html)

[oc 3.11.0](https://docs.okd.io/latest/cli_reference/get_started_cli.html#installing-the-cli)

[KVM Hypervisor](https://www.linux-kvm.org/page/Downloads)

#### Install Minishift

Make sure you have all prerequisites installed. Please check the list [here](https://docs.openshift.org/latest/minishift/getting-started/installing.html#install-prerequisites)

Download and put `minishift` in your $PATH by following steps [here](https://docs.openshift.org/latest/minishift/getting-started/installing.html#manually)

Verify installation by running following command, you should get version number.
```bash
minishift version
```

#### Install oc
Please install and set up oc on your machine by visiting [oc](https://docs.okd.io/latest/cli_reference/get_started_cli.html#installing-the-cli)

Verify installation by running following command, you should get version number.
```bash
oc version
```

### Usage
#### TL;DR
```
make minishift-start
eval $(minishift docker-env)
make deploy-all
```

Make sure that you have minishift running. you can check it using `minishift status`. If not then start it using `make minishift-start` target.

After successfully starting minishift, configure your shell to use docker daemon from minishift using `eval $(minishift docker-env)`.

Now it's time to run `git-service` and it's required resources from `deploy/resources` on OpenShift use following command:
```
make deploy-all
```

This build uses the `system:admin` account for creating all required resources from `deploy/resources` directory.

The above command then create deployment for git-service on Openshift.

#### Start Minishift
We have make target defined to start minishift with required cpu's and configuration.
```bash
make minishift-start
```

#### Exposing Docker Env
Make sure to expose minishift docker env using following command:
```bash
eval $(minishift docker-env)
```

After running this, you can verify all containers running container inside minishift using `docker ps`.

**Note**: If you miss this, docker daemon inside minishift couldn't find latest built image which causes `ImagePullBackOff` as we are using `ImagePullPolicy: IfNotPresent`


### Creating resources required to run Git Service
To create required resources to run Git Service, we have following make target:
```bash
make create-resources
```

#### Build and Deploy Git Service
To build docker image and create deployment you can use following target:
```bash
make deploy-apiserver
```

#### Creating Resources and Deploying Git Service
If you are too lazy, we have following target for you to create resources, build container and deploy service:

```bash
make deploy-all
```

#### Cleaning Up

##### Cleaning Git Service
This removes git service.
```bash
make clean-apiserver
```

##### Cleaning Resources Created to run Git Service
This removes all resources created to run git service.
```bash
make clean-resources
```

##### Cleaning Resources and Git Service
This removes all resources created to run git service along with git service itself.
```bash
make clean-all
```

#### ReDeploying Git Service with changes in code
However if you are working on git service code and wants to redeploy latest code change by building container with latest binary. We have
special target for it which will do that for you.

It won't create all other existing resources again. It'll build container and deploy git service only.

```bash
make deploy-apiserver
```