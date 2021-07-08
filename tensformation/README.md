# tensformation 
`tensformation` is a server that accepts [cloudevents](https://cloudevents.io/) from
a [Triggermesh AWS S3 Event Source](https://docs.triggermesh.io/sources/awss3/) and performs the following actions:

1. Downloads the new file in the event.
2. Base64 encodes the file.
3. Creates a request to a [Tensorflow Infrence server](https://github.com/mylesagray/docker-tensorflow-s3) instance. 
4. Emits a [cloudevent](https://cloudevents.io/) of type `io.triggermesh.transformations.tensformation.response` containing the response from the Tensorflow Infrence server as the payload. 
5. On error, `tensformation` will emit events of type `io.triggermesh.transformations.tensformation.response.error` 



## Prerequisites 
* AWS Access key
* AWS Secret key
* A [Tensorflow Infrence server](https://github.com/mylesagray/docker-tensorflow-s3) instance. 

## Building `tensformation`

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

### Running in a local enviorment
Export the required environment variables  

```
export AWS_ACCESS_KEY=
export AWS_SECRET_KEY=
export AWS_REGION=us-west-2
export TENSORFLOW_ENDPOINT=http://localhost:8501/v1/models/anpr:predict
```

From the `/tensformation` directory, run the server:
```
go run . 
```

`tensformation` expects a request similar to the following: 
```
curl -v “https://tensformation.ttst.k.triggermesh.io” \
       -X POST \
       -H “Ce-Id: 536808d3-88be-4077-9d7a-a3f162705f79" \
       -H “Ce-Specversion: 1.0” \
       -H “Ce-Type: com.amazon.s3.objectcreated” \
       -H “Ce-Source: dev.knative.samples/helloworldsource” \
       -H “Content-Type: application/json” \
       -d '  {
    “awsRegion”: “us-west-2",
    “eventName”: “ObjectCreated:Put”,
    “eventSource”: “aws:s3",
    “eventTime”: “2021-07-07T13:28:21.524Z”,
    “eventVersion”: “2.1",
    “requestParameters”: {
      “sourceIPAddress”: “162.247.91.133”
    },
    “responseElements”: {
      “x-amz-id-2”: “jlxEx2aJTL/L+XG3FFf5t6wq3FRNst07pZfdAaFWnYb8G6CVwdNeFS9DLwjmozEWxTyX/YHjzMf3jegY8V693y36lNmb8oc1",
      “x-amz-request-id”: “J65PZJJPHMHVSYYD”
    },
    “s3”: {
      “bucket”: {
        “arn”: “arn:aws:s3:::demobkt-triggermesh”,
        “name”: “demobkt-triggermesh”,
        “ownerIdentity”: {
          “principalId”: “A3L2KFRRF0JY9H”
        }
      },
      “configurationId”: “io.triggermesh.awss3sources.dmo.my-bucket”,
      “object”: {
        “eTag”: “d69281f54535f2b0ac226a3fd66916a1",
        “key”: “00000018_kuj4a4e_1.jpg”,
        “sequencer”: “0060E5ABF789917970",
        “size”: 1142782
      },
      “s3SchemaVersion”: “1.0”
    },
    “userIdentity”: {
      “principalId”: “A3L2KFRRF0JY9H”
    }
  }'
```