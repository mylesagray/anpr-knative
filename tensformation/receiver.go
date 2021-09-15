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
	response             = "io.triggermesh.transformations.s3-tensorflow.response"
)

func (recv *Receiver) receive(ctx context.Context, e cloudevents.Event) *cloudevents.Event {
	log.Printf("Processing event from source %q", e.Source())
	if typ := e.Type(); typ != s3ObjectCreatedEvent {
		fmt.Println("wrong event type")
		return emitErrorEvent("wrong event type", "wrongEventType")
	}

	req := &S3Event{}
	if err := e.DataAs(&req); err != nil {
		log.Print(err)
		return emitErrorEvent(err.Error(), "unmarshalingEvent")
	}

	image, err := recv.downloadFromS3Bucket(req)
	if err != nil {
		log.Print(err)
		return emitErrorEvent(err.Error(), "downloadingFromS3")
	}

	err, tfResponse := recv.makeTensorflowRequest(image)
	if err != nil {
		log.Print(err)
		return emitErrorEvent(err.Error(), "requestingFromTensorflow")
	}

	url := "https://" + req.S3.Bucket.Name + ".s3." + recv.region + ".amazonaws.com/" + req.S3.Object.Key

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(response)
	event.SetSource(url)
	event.SetTime(time.Now())
	err = event.SetData(cloudevents.ApplicationJSON, tfResponse)
	if err != nil {
		log.Print(err)
		return emitErrorEvent(err.Error(), "settingCEData")
	}

	return &event
}

func emitErrorEvent(er string, source string) *cloudevents.Event {
	responseEvent := cloudevents.NewEvent(cloudevents.VersionV1)
	responseEvent.SetType(response + ".error")
	responseEvent.SetSource(source)
	responseEvent.SetTime(time.Now())
	err := responseEvent.SetData(cloudevents.ApplicationJSON, er)
	if err != nil {
		log.Print(err)
		return nil
	}

	return &responseEvent

}

// downloadFromS3Bucket returns a base64 encoded string of the new image at s3
func (recv *Receiver) downloadFromS3Bucket(e *S3Event) (string, error) {
	bucket := e.S3.Bucket.Name
	item := e.S3.Object.Key

	file, err := os.Create(item)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer file.Close()

	numBytes, err := recv.s3d.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	ef := encodeFile(file)

	return ef, nil
}

func encodeFile(f *os.File) string {
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)
	encoded := base64.StdEncoding.EncodeToString(content)

	return encoded
}

func (recv *Receiver) makeTensorflowRequest(image string) (error, []byte) {
	reqBody := &TensorflowRequest{
		Instances: []struct {
			B64 string "json:\"b64\""
		}{{B64: image}},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return err, b
	}

	request, err := http.NewRequest(http.MethodPost, recv.tfEndpoint, bytes.NewBuffer(b))
	if err != nil {
		return err, b
	}

	res, err := recv.httpClient.Do(request)
	if err != nil {
		return err, b
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err, b
	}

	return nil, body
}

// TODO
// func (recv *Receiver) deleteLocalFile(){}

// func (recv *Receiver) craftCe(msg, id string) (*cloudevents.Event, error) {
// 	event := cloudevents.NewEvent(cloudevents.VersionV1)
// 	event.SetType(response)
// 	event.SetSource(id)
// 	event.SetTime(time.Now())
// 	err := event.SetData(cloudevents.ApplicationJSON, msg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &event, nil
// }
