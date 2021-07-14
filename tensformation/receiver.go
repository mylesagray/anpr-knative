/*
Copyright (c) 2021 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	s3ObjectCreatedEvent = "com.amazon.s3.objectcreated"
	response             = "io.triggermesh.transformations.tensformation.response"
)

func (recv *Receiver) receive(ctx context.Context, e cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {

	go func() {
		req := &S3Event{}
		if err := e.DataAs(&req); err != nil {
			log.Printf("unmarshaling Event: %v", err)
			recv.emitErrorEvent(err.Error(), "unmarshalingEvent")
		}

		image, err := recv.downloadFromS3Bucket(req)
		if err != nil {
			log.Printf("downloading from s3: %v", err)
			recv.emitErrorEvent(err.Error(), "downloadingFromS3")
		}

		tfResponse, err := recv.makeTensorflowRequest(image)
		if err != nil {
			log.Printf("requesting From Tensorflow: %v", err)
			recv.emitErrorEvent(err.Error(), "requestingFromTensorflow")
		}

		ceCtx := cloudevents.ContextWithTarget(context.Background(), recv.kSink)

		url := "https://" + req.S3.Bucket.Name + ".s3." + recv.region + ".amazonaws.com/" + req.S3.Object.Key
		event := cloudevents.NewEvent(cloudevents.VersionV1)
		event.SetType(response)
		event.SetSource(url)
		event.SetTime(time.Now())
		err = event.SetData(cloudevents.ApplicationJSON, tfResponse)
		if err != nil {
			log.Printf("setting cloudevent data: %v", err)
			recv.emitErrorEvent(err.Error(), "settingCEData")
		}

		if result := recv.ceClient.Send(ceCtx, event); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send, %v", result)
		}
	}()

	return nil, cloudevents.ResultACK
}

// downloadFromS3Bucket returns a base64 encoded string of the new image at s3
func (recv *Receiver) downloadFromS3Bucket(e *S3Event) (string, error) {
	bucket := e.S3.Bucket.Name
	item := e.S3.Object.Key

	file, err := os.Create(item)
	if err != nil {
		return "", err
	}
	defer file.Close()

	numBytes, err := recv.s3d.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})

	if err != nil {
		return "", err
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
	ef, err := encodeFile(file)
	if err != nil {
		return "", err
	}

	return ef, nil
}

func encodeFile(f *os.File) (string, error) {
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)
	encoded := base64.StdEncoding.EncodeToString(content)
	err := os.Remove(f.Name())
	if err != nil {
		return "", err
	}

	return encoded, nil
}

func (recv *Receiver) makeTensorflowRequest(image string) ([]byte, error) {
	reqBody := &TensorflowRequest{
		Instances: []struct {
			B64 string "json:\"b64\""
		}{{B64: image}},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, recv.tfEndpoint, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	res, err := recv.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non 200 status code returned from tensorflow server, %v", res.StatusCode)
	}

	return body, err
}

func (recv *Receiver) emitErrorEvent(er string, source string) {
	ctx := cloudevents.ContextWithTarget(context.Background(), recv.kSink)

	responseEvent := cloudevents.NewEvent(cloudevents.VersionV1)
	responseEvent.SetType(response + ".error")
	responseEvent.SetSource(source)
	responseEvent.SetTime(time.Now())
	err := responseEvent.SetData(cloudevents.ApplicationJSON, er)
	if err != nil {
		log.Printf("setting error cloudevent data: %v", err)
	}

	if result := recv.ceClient.Send(ctx, responseEvent); cloudevents.IsUndelivered(result) {
		log.Printf("failed to send, %v", result)
	}

	fmt.Println("sent event")
}
