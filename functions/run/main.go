package main

import (
  "os"
  "encoding/json"
  "github.com/apex/go-apex"
  opt "../../optimizer"
)

var debug bool = true

func main() {
  apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {

    // Marshal the event to json and log to stderr
    err := logRequest(event)
    if err != nil {
      return nil, err
    }

    // Parse the raw json
    s3Event, err := opt.ParseRequest(event)
    if err != nil {
      return err, err
    }
    
    // // Do all the stuff
    err = group.Run(s3Event)

    return err, err
  })
}

// Log the request for debugging
func logRequest(ev json.RawMessage) error {
  if debug != true {
    return nil
  }
  b, err := ev.MarshalJSON()
  if err != nil {
    return err
  }
  os.Stderr.Write(b)
  return nil
}
