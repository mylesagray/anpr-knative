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
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	sink                  = "K_SINK"
	envAccKey             = "AWS_ACCESS_KEY"
	envSecKey             = "AWS_SECRET_KEY"
	envRegion             = "AWS_REGION"
	envTensorflowEndpoint = "TENSORFLOW_ENDPOINT"
)

// Receiver runs a CloudEvents receiver.
type Receiver struct {
	s3d        *s3manager.Downloader
	kSink      string
	tfEndpoint string
	region     string
	httpClient *http.Client

	ceClient cloudevents.Client
}

func main() {
	accKey := os.Getenv(envAccKey)
	if accKey == "" {
		log.Fatal("Undefined environment variable: " + envAccKey)
	}

	kSink := os.Getenv(sink)
	if kSink == "" {
		log.Fatal("Undefined environment variable: " + sink)
	}

	region := os.Getenv(envRegion)
	if region == "" {
		log.Fatal("Undefined environment variable: " + envRegion)
	}

	tfEndpoint := os.Getenv(envTensorflowEndpoint)
	if tfEndpoint == "" {
		log.Fatal("Undefined environment variable: " + envTensorflowEndpoint)
	}

	sess, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
	downloader := s3manager.NewDownloader(sess)

	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	r := Receiver{
		s3d:        downloader,
		tfEndpoint: tfEndpoint,
		region:     region,
		kSink:      kSink,
		httpClient: http.DefaultClient,

		ceClient: c,
	}

	log.Print("Running CloudEvents receiver")

	ctx := context.Background()
	if err := r.ceClient.StartReceiver(ctx, r.receive); err != nil {
		log.Fatal(err)
	}
}
