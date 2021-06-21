from __future__ import print_function
import base64
import argparse
import requests
import numpy as np
from imutils import paths
from object_detection.utils import label_map_util
from base2designs.plates.plateFinder import PlateFinder

# Silence TF deprecation warnings in output
import logging
import os
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'  # FATAL
logging.getLogger('tensorflow').setLevel(logging.FATAL)

# Process the images
def main(server, labels_path, imagePath):
	images = paths.list_images(imagePath)
	# Loop over all the images
	for image in images:
		# infer image and get prediction from server
		with open(image, "rb") as image_file:
			jpeg_bytes = base64.b64encode(image_file.read()).decode('utf-8')
			predict_request = '{"instances" : [{"b64": "%s"}]}' % jpeg_bytes
			response = requests.post(server, data=predict_request)
		response.raise_for_status()
		prediction_classes = response.json()['predictions'][0]['detection_classes']
		prediction_scores = response.json()['predictions'][0]['detection_scores']
		prediction_boxes = response.json()['predictions'][0]['detection_boxes']
		prediction_num_detections = response.json()['predictions'][0]['num_detections']

		# squeeze the lists into a single dimension
		boxes = np.squeeze(prediction_boxes)
		scores = np.squeeze(prediction_scores)
		labels = np.squeeze(prediction_classes)

		# load the class labels from disk
		labelMap = label_map_util.load_labelmap(labels_path)
		categories = label_map_util.convert_label_map_to_categories(
			labelMap, max_num_classes=37,
			use_display_name=True)
		categoryIdx = label_map_util.create_category_index(categories)

		# create a plateFinder
		plateFinder = PlateFinder(0.5, categoryIdx,
								rejectPlates=False, charIOUMax=0.3)

		# Perform inference on the full image, and then find the plate text associated with each plate
		licensePlateFound_pred, plateBoxes_pred, charTexts_pred, charBoxes_pred, charScores_pred, plateScores_pred = plateFinder.findPlates(
			boxes, scores, labels)

		# Print plate text
		for charText in charTexts_pred:
			print("    Found plate: ", charText)

if __name__ == '__main__':
  ap = argparse.ArgumentParser()
  ap.add_argument("-s", "--server", required=True,
                  help="TF Serving service url")
  ap.add_argument("-l", "--labels", required=True,
                  help="labels file")
  ap.add_argument("-i", "--imagePath", required=True,
                  help="path to input image path")

  args = vars(ap.parse_args())
  results = main(args["server"], args["labels"], args["imagePath"])
