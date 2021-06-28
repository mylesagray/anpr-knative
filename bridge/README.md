# Tensorflow-s3-docker
## Deploying the bridge

1: Deploy the custom tensorflow server in the namespace.

```
kn service create tf-inference-server -n default --autoscale-window 300s \
  --request "memory=2Gi" \
  -p 8501 --image harbor-repo.vmware.com/vspheretmm/anpr-serving \
  --arg --model_config_file=/configs/models-local.config \
  --arg --monitoring_config_file=/configs/monitoring_config.txt
```

2: Update the `manifest.yaml` file, replacing the placeholder `""` marks.

3: Deploy the bridge.
```
kubectl -n default apply -f manifest.yaml
```

## Building the containers
### Building tensformation
1: Move to the tensformation directory.
```
cd tensformation
```

2: Create the go.mod file.
```
go mod init
```

3: Build & submit the dockerfile.
```
gcloud builds submit --tag gcr.io/<project>/tensformation .
```

### Building tensorflow_client
1: Move to the `tensorflow_client` directory.
```
cd tensorflow_client
```

2: Build & submit the dockerfile.
```
gcloud builds submit --tag gcr.io/<project>/tfclient .
```