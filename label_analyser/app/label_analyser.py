# Copyright (c) 2021 TriggerMesh Inc.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
from __future__ import print_function
from flask import Flask, request, make_response, jsonify
import uuid
import base64
import argparse
import requests
import numpy as np
from imutils import paths
from base2designs.plates.plateFinder import PlateFinder
from cloudevents.sdk import converters
from cloudevents.sdk import marshaller
from cloudevents.sdk.converters import structured
from cloudevents.http import CloudEvent, to_structured
from cloudevents.sdk.event import v1
from object_detection.utils import label_map_util
from object_detection.protos import string_int_label_map_pb2
from google.protobuf import text_format
from apiclient import discovery
from google.oauth2 import service_account
from datetime import datetime
from cloudevents.http import from_http

import logging
import os
import json

os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'  # FATAL
logging.getLogger('tensorflow').setLevel(logging.FATAL)

app = Flask(__name__)

sink = os.environ['K_SINK']


def _validate_label_map(label_map):
  """Checks if a label map is valid.
  Args:
    label_map: StringIntLabelMap to validate.
  Raises:
    ValueError: if label map is invalid.
  """
  for item in label_map.item:
    if item.id < 0:
      raise ValueError('Label map ids should be >= 0.')
    if (item.id == 0 and item.name != 'background' and
        item.display_name != 'background'):
      raise ValueError('Label map id 0 is reserved for the background label')

def load_labelmap(path):
  """Loads label map proto.

  Args:
    path: path to StringIntLabelMap proto text file.
  Returns:
    a StringIntLabelMapProto
  """
  with open(path, 'r') as fid:
    label_map_string = fid.read()
    label_map = string_int_label_map_pb2.StringIntLabelMap()
    try:
      text_format.Merge(label_map_string, label_map)
    except text_format.ParseError:
      label_map.ParseFromString(label_map_string)
  _validate_label_map(label_map)
  return label_map

def emitErrorEvent(err, source):
    attributes = {
    "type": "io.triggermesh.functions.tensorflow.label.analyser.response.error",
    "source": source,
    }
    data = { "error": err}

    event = CloudEvent(attributes, data)
    headers, body = to_structured(event)

    # send and print event
    requests.post(sink, headers=headers, data=body)
    print(f"Sent {event['id']} from {event['source']} with " f"{event.data}")
    return 

def emitNoTagFoundEvent(data, source):
    attributes = {
    "type": "io.triggermesh.functions.tensorflow.label.analyser.response.noid",
    "source": source,
    }

    event = CloudEvent(attributes, data)
    headers, body = to_structured(event)

    # send and print event
    requests.post(sink, headers=headers, data=body)
    print(f"Sent {event['id']} from {event['source']} with " f"{event.data}")
    return 

@app.route('/', methods=['POST'])
def hello_world():
    try:
      req_data = request.get_json()
      prediction_classes = req_data['predictions'][0]['detection_classes']
      prediction_scores = req_data['predictions'][0]['detection_scores']
      prediction_boxes = req_data['predictions'][0]['detection_boxes']
      boxes = np.squeeze(prediction_boxes)
      scores = np.squeeze(prediction_scores)
      labels = np.squeeze(prediction_classes)

      # load the class labels from disk
      labelMap = load_labelmap("classes.pbtxt")
      categories = label_map_util.convert_label_map_to_categories(labelMap, max_num_classes=37,use_display_name=True)
      categoryIdx = label_map_util.create_category_index(categories)
    
      # create a plateFinder
      plateFinder = PlateFinder(0.5, categoryIdx,rejectPlates=False, charIOUMax=0.3)

      # Perform inference on the full image, and then find the plate text associated with each plate
      licensePlateFound_pred, plateBoxes_pred, charTexts_pred, charBoxes_pred, charScores_pred, plateScores_pred = plateFinder.findPlates(boxes, scores, labels)

      event = from_http(request.headers, request.get_data())
      imageURL = event['source']

      # Print plate text
      foundPlate = ""
      for charText in charTexts_pred:
          print("Found plate: ", charText)
          foundPlate = charText
      if foundPlate != "":
        attributes = {
            "type": "io.triggermesh.functions.tensorflow.label.analyser.response.plateid",
            "source": "tfclient",
        }
        data = { "plate": foundPlate, "url": imageURL}

        event = CloudEvent(attributes, data)
        headers, body = to_structured(event)

        # send and print event
        requests.post(sink, headers=headers, data=body)
        # print(f"Sent {event['id']} from {event['source']} with " f"{event.data}")

        creds = os.environ['GOOGLE_CREDENTIALS_JSON']
        spreadsheet_id = os.environ['SHEET_ID']

        scopes = ["https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/drive.file", "https://www.googleapis.com/auth/spreadsheets"]
        credentials = service_account.Credentials.from_service_account_info(json.loads(creds), scopes=scopes)
        service = discovery.build('sheets', 'v4', credentials=credentials)

        now = datetime.now()
        time_now = now.strftime("%d/%m/%Y, %H:%M:%S")

        rows = [
            [foundPlate, imageURL, time_now ],
        ]

        result = service.spreadsheets().values().get(spreadsheetId=spreadsheet_id,range="Sheet1!A:Z").execute()

        values = result.get('values')

        lr = len(values)

        service.spreadsheets().values().append(spreadsheetId=spreadsheet_id,range="Sheet1!A"+str(lr)+":Z",body={"majorDimension": "ROWS","values": rows},valueInputOption="USER_ENTERED").execute()

        return data

      if foundPlate == "":
        event = from_http(request.headers, request.get_data())
        headers, body = to_structured(event)
        emitNoTagFoundEvent(event.data,"noPlateID")
        return 'no plate match'

    except KeyError as e:
      emitErrorEvent(str(e) ,"keyError")
      return str(e)

    except TypeError as e:
      emitErrorEvent(str(e),"typeError")
      return str(e)
      
    except FileNotFoundError as e:
     emitErrorEvent(str(e),"fileNotFoundError")
     return str(e)

    except Exception as e:
     print("Oops!", e, "occurred.")
     emitErrorEvent( str(e),"general")
     return str(e)

    return 'ok'



if __name__ == '__main__':
    app.debug = True
    app.run(host='0.0.0.0', port=8080)
