package optimizer

import (
  "os"
  "strings"
  "fmt"
  "encoding/json"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/s3"
  "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Event struct {
	Records []Record `json:"Records"`
}

type Record struct {
  EventName string `json:"eventName"`
  S3        S3     `json:"s3"`
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

type Group struct {
  Bucket     string
  Prefix     string
  Directives []Directive
}

type Groups []Group

/*
  TODO: Command should be new type
  struct {
    Head string
    Arguments []string
  }
*/
type Directive struct {
  File    string
  Glob    []string
  Command string
}

// TODO: use this instead of string for Directive.Command
type Command struct {
  Head      string
  Arguments string
}

type keyparts struct {
  path      string
  slug      string
  extension string
}

var (
  sess          *session.Session
  tmpPath       string
  localOriginal string
  groups        *Groups
)

func init() {

  // load the config files
  groups = loadGroups()

  // Initialize S3 session
  sess = session.Must(session.NewSession(&aws.Config{
  	Region: aws.String("us-east-1"),
  }))
}

// Run all the stuff
// TODO: break out create stuff and instead use a switch case to route the request
func Run(ev *S3Event) error {
  group, err := getGroup(ev, groups)
  if err != nil {
    return err
  }
  if ev.isCreate() {
    return ev.create(group)
  } else if ev.isDestroy() {
    return ev.destroy(group)
  } else {
    return fmt.Errorf("%s is unrecongized event type.", ev.Records[0].EventName)
  }
}

// TODO
func (ev *S3Event) isCreate() bool {
  return true
}

// TODO
func (ev *S3Event) isDestroy() bool {
  return false
}

// TODO: figure out which group to use based on request
func getGroup(se *S3Event, g *Groups) (*Group, error) {
  return &Group{}, nil
}

// Separate the extension from the original key into path/slug/extension
// e.g., /this/is/a/file.jpg -> /this/is/a, file, jpg
func newKeyParts(in string) *keyparts {
  pathParts := strings.Split(in, ".")
  ext := pathParts[len(pathParts)-1]
  pathParts = pathParts[:len(pathParts)-1]
  path := strings.Join(pathParts, ".")
  pathParts = strings.Split(path, "/")
  slug := pathParts[len(pathParts)-1]
  pathParts = pathParts[:len(pathParts)-1]
  return &keyparts {strings.Join(pathParts, "/"), slug, ext}
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

// Download
func download(localTarget, bucket, key string) error {
  f, err := os.Create(localTarget)
  if err != nil {
    return nil
  }
  defer f.Close()
  downloader := s3manager.NewDownloader(sess)
  _, err = downloader.Download(f, &s3.GetObjectInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(key),
  })
  return err
}

// Upload to s3 from local file
func upload(localPath, bucket, key string) error {
  f, err := os.Open(localPath)
  if err != nil {
    return err
  }
  defer f.Close()
  uploader := s3manager.NewUploader(sess)
  resp, err := uploader.Upload(&s3manager.UploadInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(key),
    Body:   f,
  })
  fmt.Println(resp)
  return err
}

// TODO
func batchUpload(files []*os.File) error {
  return nil
}

// Delete a media object
func delete() error {
  return nil
}

// TODO: load config files into memory
func loadGroups() *Groups {
  return &Groups{}
}

// TODO: intilializer for Group
func newGroup() *Group {
  return &Group{}
}

// TODO: validate group
func (g *Group) validate() (bool, error) {
  return false, nil
}

// TODO: initializer for directive
func newDirective() *Directive {
  return &Directive{}
}

/*
 TODO: validate directive
 - Ensure unique ID
 */
func (d *Directive) validate() (bool, error) {
  return false, nil
}
