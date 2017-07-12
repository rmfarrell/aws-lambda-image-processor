package optimizer

import (
  "encoding/json"
)

type S3Event struct {
	Records []Record `json:"Records"`
}

type Record struct {
  EventName string    `json:"eventName"`
  S3        S3        `json:"s3"`
}

type S3 struct {
  Bucket Bucket `json:"bucket"`
  Object Object `json:"object"`
}

type Object struct {
  Key  string `json:"key"`
  Size int    `json:"size"`
}

type Bucket struct {
  Name string `json:"name"`
  Arn  string `json:"arn"`
}

type Config struct {
  Destination string
  MimeType    string
  Matcher     string
  Actions     []string
}

type Projects []Config

// Load the configuration objects
func init() {}

// Run all the stuff
func HandleS3Event(in json.RawMessage) (*S3Event, error) {
  s3Event, err := parseRequest(in)
  if err != nil {
    return nil, err
  }

  return s3Event, err
}

// Parse an S3 event JSON -> S3event
func parseRequest(in json.RawMessage) (*S3Event, error) {
  se := S3Event{}
  err := json.Unmarshal(in, &se)
  if err != nil {
    return nil, err
  }
  return &se, nil
}

// Create a media object
func create() error {
  return nil
}

// Delete a media object
func delete() error {
  return nil
}
