from __future__ import print_function
from flask import Flask
from flask import request
import base64
import argparse
import requests
import numpy as np
from imutils import paths
from object_detection.utils import label_map_util
from base2designs.plates.plateFinder import PlateFinder
from cloudevents.sdk import converters
from cloudevents.sdk import marshaller
from cloudevents.sdk.converters import structured
from cloudevents.http import CloudEvent, to_structured
from cloudevents.sdk.event import v1


# Silence TF deprecation warnings in output
import logging
import os
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'  # FATAL
logging.getLogger('tensorflow').setLevel(logging.FATAL)

app = Flask(__name__)

@app.route('/', methods=['POST'])
def hello_world():
    sink = os.environ['K_SINK']
    req_data = request.get_json()
    # print(req_data)
    prediction_classes = req_data['predictions'][0]['detection_classes']
    prediction_scores = req_data['predictions'][0]['detection_scores']
    prediction_boxes = req_data['predictions'][0]['detection_boxes']
    boxes = np.squeeze(prediction_boxes)
    scores = np.squeeze(prediction_scores)
    labels = np.squeeze(prediction_classes)

    # load the class labels from disk
    labelMap = label_map_util.load_labelmap("classes.pbtxt")
    categories = label_map_util.convert_label_map_to_categories(labelMap, max_num_classes=37,use_display_name=True)
    categoryIdx = label_map_util.create_category_index(categories)

    # # create a plateFinder
    plateFinder = PlateFinder(0.5, categoryIdx,rejectPlates=False, charIOUMax=0.3)

    # Perform inference on the full image, and then find the plate text associated with each plate
    licensePlateFound_pred, plateBoxes_pred, charTexts_pred, charBoxes_pred, charScores_pred, plateScores_pred = plateFinder.findPlates(boxes, scores, labels)

    # Print plate text
    x = ""
    for charText in charTexts_pred:
        print("Found plate: ", charText)
        x = charText

    attributes = {
        "type": "io.triggermesh.functions.tensorflow.client",
        "source": "https://example.com/event-producer",
    }
    data = {"plate": x}
    event = CloudEvent(attributes, data)
    
    # Creates the HTTP request representation of the CloudEvent in structured content mode
    headers, body = to_structured(event)

    requests.post(sink, data=body, headers=headers)

    return (x)

if __name__ == "__main__":
   app.run(debug=True,host='0.0.0.0',port=int(os.environ.get('PORT', 8080)))