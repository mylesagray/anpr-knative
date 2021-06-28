/*
Copyright (c) 2020 TriggerMesh Inc.

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

import "time"

// S3Event is the structure of the event we expect to receive.
type S3Event struct {
	Awsregion         string    `json:"awsRegion"`
	Eventname         string    `json:"eventName"`
	Eventsource       string    `json:"eventSource"`
	Eventtime         time.Time `json:"eventTime"`
	Eventversion      string    `json:"eventVersion"`
	Requestparameters struct {
		Sourceipaddress string `json:"sourceIPAddress"`
	} `json:"requestParameters"`
	Responseelements struct {
		XAmzID2       string `json:"x-amz-id-2"`
		XAmzRequestID string `json:"x-amz-request-id"`
	} `json:"responseElements"`
	S3 struct {
		Bucket struct {
			Arn           string `json:"arn"`
			Name          string `json:"name"`
			Owneridentity struct {
				Principalid string `json:"principalId"`
			} `json:"ownerIdentity"`
		} `json:"bucket"`
		Configurationid string `json:"configurationId"`
		Object          struct {
			Etag      string `json:"eTag"`
			Key       string `json:"key"`
			Sequencer string `json:"sequencer"`
			Size      int    `json:"size"`
		} `json:"object"`
		S3Schemaversion string `json:"s3SchemaVersion"`
	} `json:"s3"`
	Useridentity struct {
		Principalid string `json:"principalId"`
	} `json:"userIdentity"`
}

type TensorflowRequest struct {
	Instances []struct {
		B64 string `json:"b64"`
	} `json:"instances"`
}

type TensorflowResponse struct {
	Predictions []struct {
		// DetectionClasses          []int       `json:"detection_classes"`
		NumDetections             float64     `json:"num_detections"`
		DetectionBoxes            [][]float64 `json:"detection_boxes"`
		RawDetectionBoxes         [][]float64 `json:"raw_detection_boxes"`
		DetectionScores           []float64   `json:"detection_scores"`
		RawDetectionScores        [][]float64 `json:"raw_detection_scores"`
		DetectionMulticlassScores [][]float64 `json:"detection_multiclass_scores"`
	} `json:"predictions"`
}
