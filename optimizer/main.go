package optimizer

import (
  "os"
  "fmt"
  "bytes"
  "strings"
  "os/exec"
  "encoding/json"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/s3"
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

type Destination struct {
  Bucket     string
  Prefix     string
  Directives []Config
}

// TODO: rename to directive???
/*
  TODO: Command should be new type
  struct {
    Head string
    Arguments []string
  }
*/
type Config struct {
  MimeTypes []string
  Matcher   string
  Command   string
  VariantId string
  Key       string
}

var (
  sess        *session.Session
  destination *Destination
)

const (
  tmpPath string = "/tmp/imgopt"
)

func init() {

  // load the config files
  destination = loadProjectsConfigs()

  // Initialize S3 session
  sess = session.Must(session.NewSession(&aws.Config{
  	Region: aws.String("us-east-1"),
  }))
}

// Run all the stuff
func Start(ev *S3Event) error {

  // Create tmp dir
  err := makeTmp()
  if err != nil {
    return err
  }

  // identifyproject. Key off `object.key` to figure out which project the
  // object belongs to

  // download file to tmp, return tmp path to file
  localSrc, err := getObject(ev)
  if err != nil {
    return err
  }

  // apply each config item to downloaded file
  for _, config := range destination.Directives {
    cmd := replaceSourceAndDestination(localSrc, &config)
    err = executeCommand(cmd)
    if err != nil {
      return err
    }

    // upload result
    // putObject()

  }

  // cleanup
  err = removeTmp()
  if err != nil {
    return err
  }

  // write manifest?

  return err
}

// Download the object from S3 into `/tmp` directory
// TODO
func getObject(s3e *S3Event) (string, error) {
  svc := s3.New(sess)
  input := &s3.GetObjectInput{
    Bucket: aws.String("examplebucket"),
    Key:    aws.String("HappyFace.jpg"),
  }

  result, err := svc.GetObject(input)
  if err != nil {
    return  "", err
  }

  fmt.Println(result)
  return "", nil
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

// Upload to s3
func putObject(localPath, bkt, key string) error {
  svc := s3.New(sess)
  _, err := svc.PutObject(&s3.PutObjectInput{
    Body:                 aws.ReadSeekCloser(strings.NewReader(localPath)),
    Bucket:               aws.String(bkt),
    Key:                  aws.String(key),
  })
  if err != nil {
    return err
  }
  return nil
}

// make the tmp staging directory
// TODO: use a uuid
func makeTmp() error {
  return os.Mkdir(tmpPath, 0777)
}

// Clean /tmp of files
func removeTmp() error {
  return os.Remove(tmpPath)
}

// Apply the action to the fle.
func executeCommand(in string) error {
  var buf bytes.Buffer

  // break the command into command and args
  parts := strings.Fields(in)
  head := parts[0]
  parts = parts[1:len(parts)]

  cmd := exec.Command(head, parts...)
	err := cmd.Start()
	if err != nil {
    fmt.Println(err)
    return err
	}
  cmd.Stderr = &buf
  cmd.Stdout = &buf
  fmt.Println(buf.String())
  return nil
}

// String replacement operation for {source} and {destination}
func replaceSourceAndDestination(src string, c *Config) string {
  cmd := strings.Replace(c.Command, "{source}", src, 1)
  return strings.Replace(cmd, "{destination}", fmt.Sprintf("%s/%s", tmpPath, c.VariantId), 1)
}

// Delete a media object
func delete() error {
  return nil
}

// TODO
func loadProjectsConfigs() *Destination {
  return &Destination{}
}
