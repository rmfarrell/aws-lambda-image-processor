package optimizer

import (
  "os"
  "os/exec"
  "bytes"
  "fmt"
  "strings"
  "github.com/satori/go.uuid"
)

func (ev *S3Event) create(group *Group) error {

  var (
    key string = ev.Records[0].S3.Object.Key
    src string = ev.Records[0].S3.Bucket.Name
    dest string = group.Destination.BucketName
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
      fmt.Sprintf("%s/%s", tmpPath, d.File),
      dest,
      fmt.Sprintf("%s%s/%s", group.Destination.Root, kp.slug, d.File))
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


// ------------------------------ Helpers ------------------------------

// String replacement operation for {source} and {destination}
func replaceSourceAndDestination(src string, c *Directive) string {
  cmd := strings.Replace(c.Command, "{source}", src, 1)
  return strings.Replace(cmd, "{destination}", fmt.Sprintf("%s/%s", tmpPath, c.File), 1)
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