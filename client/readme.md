# Using the TF client app

## Install Anaconda

```sh
brew install anaconda
echo 'PATH=$PATH:/usr/local/anaconda3/bin/' > ~/.zshrc
```

## Create conda env with requirements

```sh
conda create --name tf-1.15 --file app/requirements.txt
```

## Activate the env

```sh
conda activate tf-1.15
```

## Execute the client

```sh
python app/predict_images_client.py -s http://tf-inference-server.default.10.198.53.135.sslip.io/v1/models/anpr:predict -i ../test/cars/ -l ../test/classes.pbtxt
```
