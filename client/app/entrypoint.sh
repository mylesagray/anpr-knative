#!/bin/bash --login
# The --login ensures the bash configuration is loaded,
# enabling Conda.
set -euo pipefail
conda activate tf-1.15
exec python predict_images_client.py -s http://tf-inference-server.default.10.198.53.135.sslip.io/v1/models/anpr:predict -i test/cars/ -l test/classes.pbtxt