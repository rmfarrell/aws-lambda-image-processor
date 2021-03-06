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
  "regexp"
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
  Source      Target
  Destination Target
  Directives  []Directive
}

// A Bucket and the path in that bucket to a root
type Target struct {
  BucketName string
  Root       string
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
  sess   *session.Session
  groups *Groups
)

// Run all the stuff
// TODO: break out create stuff and instead use a switch case to route the request
// TODO: return []error instead of a single error if recoverable
func (g *Group) Run(ev *S3Event) error {

  // Initialize S3 session
  sess = createSession()

  // Route request
  if ev.isCreate() {
    return ev.create(g)
  } else if ev.isDestroy() {
    return ev.destroy(g)
  } else {
    return fmt.Errorf("%s is unrecongized event type.", ev.Records[0].EventName)
  }
}

// Parse an S3 event JSON -> S3event
func ParseRequest(in json.RawMessage) (*S3Event, error) {
  se := S3Event{}
  err := json.Unmarshal(in, &se)
  if err != nil {
    return nil, err
  }
  return &se, nil
}

// -------------------------- Private --------------------------

// Initialize an AWS session
func createSession() *session.Session {
  return session.Must(session.NewSession(&aws.Config{
    Region: aws.String("us-east-1"),
  }))
}

// Check whether event is a create Event
func (ev *S3Event) isCreate() bool {
  p := regexp.MustCompile("^ObjectCreated")
  return p.MatchString(ev.Records[0].EventName)
}

// Check whether event is a destroy Event
func (ev *S3Event) isDestroy() bool {
  p := regexp.MustCompile("^ObjectRemoved")
  return p.MatchString(ev.Records[0].EventName)
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
  _, err = uploader.Upload(&s3manager.UploadInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(key),
    Body:   f,
  })
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

// TODO: necessary for Target roots with leading slash
func stripLeadingSlash(s string) string { return s }
func stripFollowingSlash(s string) string { return s }
