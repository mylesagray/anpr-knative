# Automatic Number Plate Recognition based on KNative

An auto-scaling ML-based number plate recognition system, running on KNative.

![Demo of app inferring a number plate](img/demo.gif)

This repo takes the Automatic Number Plate Recognition (ANPR) TensorFlow container detailed here: <https://github.com/mylesagray/docker-tensorflow-s3>, packages the TensorFlow Python client built to interact with the model and adds a "tensformation" component to deploy an event-based auto-scaling ANPR system on KNative.

## Overview

![Architecture Overview](./img/overview.png)

1. This bridge starts with an event propagated by the s3 source.
<br>
</br>
1. This event is then consumed by `tensformation`, this service then performs the following actions:
    * Downloads the file.
    * Base64 encodes it.
    * Creates a request to the `Tensorflow Server`.
    * Returns an event containing the `Tensorflow Server` ANPR model's raw response.
<br>
</br>
1. The found plate event is then consumed by the `label_analyser`, this service then performs the following actions:
    * Performs the Tensorflow response analysis on the output from the model.
    * Updates the provided Google Sheet with the found plate info.
    * Returns an event back to the broker with the found plate.
<br>
</br>
1. Depending upon the outcome of `lable_analyser` the following action(s) are performed:

    **If there is no plate found within the image:**
    * A Datadog statistic is updated to be viewed in the dashboard. 
    * A Slack message is posted in a designated channel conatining a similar message:

        `Image failed to process: https://tmdmobkt.s3.us-west-2.amazonaws.com/pika.jpg`
    If there is a plate found within the image:
    * This Google sheet is updated -> https://docs.google.com/spreadsheets/d/1NIJOyekYYYGmu1sBMKgnDse68sgNynyJbfEgKv4UWMU/edit#gid=0
    * This Firestore collection is updated -> https://console.cloud.google.com/firestore/data/plate_id/urwe4l2?project=ultra-hologram-297914

    **If there is a bad guy detected:**
    * An SMS message is sent to the Google voice account linked to the `demo@triggermesh.com` 
    * This Google sheet is updated -> https://docs.google.com/spreadsheets/d/1NIJOyekYYYGmu1sBMKgnDse68sgNynyJbfEgKv4UWMU/edit#gid=0
    * This Firestore collection is updated -> https://console.cloud.google.com/firestore/data/plate_id/urwe4l2?project=ultra-hologram-297914

   **If an error occurs:**
    * A Zendesk ticket is created containing the service that failed in the title and the error in the body of the ticket. 


## Deploying the application

### Prerequisites

#### Google

You need to set up a Google Cloud Project and service user to use the Google Sheets output, please follow the following two sections of documentation:

* <https://docs.triggermesh.io/targets/googlesheets/#google-api-credentials>
* <https://docs.triggermesh.io/targets/googlesheets/#googlesheets-sheet-id>

You will need both the JSON key output and the Google Sheet ID for this app to run, additionally - don't forget to share the Google Sheet you want to use as the target for this app with the service account email address created above.

#### Amazon

You will need to create an S3 bucket with public access enabled and get its ARN (take particular note in the below docs on where and how to add your region and account ID to the ARN provided by AWS):

* <https://docs.triggermesh.io/sources/awss3/#amazon-resource-name-arn>

Additionally, you will need to create or get your AWS Access Key and Secret:

* <https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html#Using_CreateAccessKey>

#### Knative and TriggerMesh

It is required to run Knative and the TriggerMesh sources for this demo, you can find details on how to do that below:

* <https://docs.vmware.com/en/Cloud-Native-Runtimes-for-VMware-Tanzu/1.0/tanzu-cloud-native-runtimes-1-0/GUID-cnr-overview.html>

You will need a Knative broker, we don't include one in the manifest as these are generally quite opinionated, so if you just want to get it up and running use the default broker:

```sh
kn broker create default
```

If you list the brokers, you'll find the URL for the broker you just created - just plug this into `manifest.yaml` under `K_SINK`:

```sh
$ kn broker list
NAME      URL                                                                        AGE     CONDITIONS   READY   REASON
default   http://broker-ingress.knative-eventing.svc.cluster.local/default/default   7d22h   5 OK / 5     True    
```

### Deploy the app

1: Update the `manifest.yaml` file, replacing the placeholder `''` marks with your information.

2: Deploy the app.

```sh
kubectl -n default apply -f manifest.yaml
```

## Running the app

Drop an image of a car or vehicle with a US number plate into the S3 bucket targeted in the deployment above and watch the Google Sheet as a row containing the number plate, image URL and timestamp are populated.
## Building the containers

### Building `tensformation`

1: Move to the `tensformation` directory.

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

### Building `label_analyser`

1: Move to the `label_analyser` directory.

```sh
cd label_analyser
```

2: Build & submit the dockerfile.

```sh
docker build . -t harbor-repo.vmware.com/vspheretmm/tfclient:latest -t tfclient:latest
docker push harbor-repo.vmware.com/vspheretmm/tfclient:latest

## OR

gcloud builds submit --tag gcr.io/<project>/tfclient .
```
