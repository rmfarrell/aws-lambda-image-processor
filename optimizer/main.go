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
  "github.com/satori/go.uuid"
  "github.com/aws/aws-sdk-go/service/s3/s3manager"
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

type Group struct {
  Bucket     string
  Prefix     string
  Directives []Directive
}

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
  group         *Group
  tmpPath       string
  localOriginal string
)

func init() {

  // load the config files
  group = loadGroups()

  // Initialize S3 session
  sess = session.Must(session.NewSession(&aws.Config{
  	Region: aws.String("us-east-1"),
  }))
}

// Run all the stuff
// TODO: break out create stuff and instead use a switch case to route the request
func Start(ev *S3Event) error {

  var (
    key string = ev.Records[0].S3.Object.Key
    src string = ev.Records[0].S3.Bucket.Name
    dest string = group.Bucket
  )

  // Create tmp dir
  t, err := makeTmp()
  if err != nil {
    return err
  }
  tmpPath = t
  localOriginal = fmt.Sprintf("%s/%s", tmpPath, key)

  // cleanup
  // defer

  // identifyproject. Key off `object.key` to figure out which project the
  // object belongs to

  err = download(localOriginal, src, key)
  if err != nil {
    return err
  }

  // apply each config item to downloaded file
  // TODO: set up channel, run concurrently
  // TODO: handle response from executeCommand
  for _, d := range group.Directives {
    cmd := replaceSourceAndDestination(localOriginal, &d)
    _, err = executeCommand(cmd)
    if err != nil {
      return err
    }

    // upload result
    kp := newKeyParts(key)
    err = upload(
      fmt.Sprintf("%s/%s", tmpPath, d.ID),
      dest,
      fmt.Sprintf("%s%s/%s", kp.path, kp.slug, d.ID))
    if err != nil {
      return err
    }
  }

  err = removeDir(tmpPath)
  if err != nil {
    return err
  }

  // write manifest?

  return err
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

// Get the objects head
func getObject(bucket, key string) (string, error) {
  svc := s3.New(sess)
  _, err := svc.GetObject(&s3.GetObjectInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(key),
  })
  if err != nil {
    return  "", err
  }
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

// make the tmp staging directory
func makeTmp() (string, error) {
  path := fmt.Sprintf("/tmp/%s", uuid.NewV4().String())
  err := os.Mkdir(path, 0777)
  if err != nil {
    return "", err
  }
  return path, nil
}

// Clean /tmp of files
func removeDir(p string) error {
  return os.RemoveAll(p)
}

// Apply the action to the fle.
func executeCommand(in string) (string, error) {
  var buf bytes.Buffer

  // break the command into command and args
  parts := strings.Fields(in)
  head := parts[0]
  parts = parts[1:len(parts)]

  cmd := exec.Command(head, parts...)
	err := cmd.Start()
	if err != nil {
    fmt.Println(err)
    return "", err
	}
  cmd.Stderr = &buf
  cmd.Stdout = &buf
  err = cmd.Wait()
  return buf.String(), err
}

// String replacement operation for {source} and {destination}
func replaceSourceAndDestination(src string, c *Directive) string {
  cmd := strings.Replace(c.Command, "{source}", src, 1)
  return strings.Replace(cmd, "{destination}", fmt.Sprintf("%s/%s", tmpPath, c.ID), 1)
}

// Delete a media object
func delete() error {
  return nil
}

// TODO: load config files into memory
func loadGroups() *Group {
  return &Group{}
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
