FROM python:3.7-slim

COPY . .

RUN apt-get update && apt-get install
RUN apt-get update && apt-get install -y python3-opencv
RUN apt-get install -y  python-dev
RUN apt-get install -y  build-essential
RUN apt-get install -y apt-utils

RUN pip install --upgrade setuptools && \
    pip install https://storage.googleapis.com/tensorflow/mac/cpu/tensorflow-1.5.0-py3-none-any.whl && \
    pip install keras --no-deps && \
    pip install opencv-python && \
    pip install imutils && \
    pip install flask && \
    pip install cloudevents && \
    pip install tf_object_detection && \
    pip install requests && \
    pip install --upgrade google-api-python-client google-auth-httplib2 google-auth-oauthlib


EXPOSE 8080

CMD ["python", "-u", "/tensorflow.py"]
