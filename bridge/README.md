# Tensorflow-s3-docker

## Overview

![ov](./img/overview.png)

This bridge starts with an event propagated by the s3 source. 

This event is then consumed by `tensformation`, this service then performs the following actions:

* Downloads the file.
* Base64 encodes it.
* Creates a request from the `Tensorflow server`.
* Returns an event containing the Tensorflow server response. 

The found plate event is then consumed by the `tensorflow_client`, this service then performs the following actions:

* Performs the Tensorflow response analysis.
* Updates the provided Google Sheet with the found plate info.
* Returns an event back to the broker with the found plate. 

An event-display log dump of the cloudevent transactions performed in this bridge can be found in the file `ce_dump.txt`

## Deploying the bridge

1: Update the `manifest.yaml` file, replacing the placeholder `""` marks with your information.

2: Deploy the bridge.

```sh
kubectl -n default apply -f manifest.yaml
```

## Building the containers

### Building tensformation

1: Move to the tensformation directory.

```sh
cd tensformation
```

2: Create the go.mod file.

```sh
go mod init tensformation
```

3: Build & submit the dockerfile.

```sh
docker build . -t harbor-repo.vmware.com/vspheretmm/tensformation:latest -t tensformation:latest
docker push harbor-repo.vmware.com/vspheretmm/tensformation:latest

## OR

gcloud builds submit --tag gcr.io/<project>/tensformation .
```

### Building tensorflow_client

1: Move to the `tensorflow_client` directory.

```sh
cd tensorflow_client
```

2: Build & submit the dockerfile.

```sh
docker build . -t harbor-repo.vmware.com/vspheretmm/tfclient:latest -t tfclient:latest
docker push harbor-repo.vmware.com/vspheretmm/tfclient:latest

## OR

gcloud builds submit --tag gcr.io/<project>/tfclient .
```
