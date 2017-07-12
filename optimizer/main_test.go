package optimizer

import (
  "testing"
  "os"
  "fmt"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/service/s3"
)

// -- Test config
const (
  sourceBucket string = "image-optimizer-bucket-src-test"
  destBucket string = "image-optimizer-bucket-dest-test"
)

func TestMain(m *testing.M) {
  setup()
  retCode := m.Run()
  // teardown()
  os.Exit(retCode)
}

func TestStart(t *testing.T) {
  destination = &Destination {
    destBucket,
    "this/is/a/path",
    []Config{
      Config {
        []string{ "image/png" },
        "*",
        "convert {source} -resize x64> {destination}",
        "converted.png",
        "",
      },
    },
  }
  err := Start(mockCreate("test.png"))
  if err != nil {
    t.Fatal(err)
  }
}


// ---------------------- Helpers ----------------------

// Create a mock s3 create event
func mockCreate(obj string) *S3Event {
  return &S3Event {
    []Record {
      Record {
        "ObjectCreated:Put",
        S3 {
          Bucket {
            "carrot-image-handler-test",
            fmt.Sprintf("arn:aws:s3:::%s", sourceBucket),
          },
          Object {
            obj,
            0,
          },
        },
      },
    },
  }
}

// Create a mock s3 delete event
func mockDelete(obj string) *S3Event {
  return &S3Event {
    []Record {
      Record {
        "ObjectRemoved:Delete",
        S3 {
          Bucket {
            "carrot-image-handler-test",
            fmt.Sprintf("arn:aws:s3:::%s", sourceBucket),
          },
          Object {
            obj,
            0,
          },
        },
      },
    },
  }
}


// ---------------------- Setup/teardown ----------------------

func setup() {
  // Create a destination bucket
  err := createBucket(destBucket)
  if err != nil {
    fmt.Println(err)
  }
  err = createBucket(sourceBucket)
  if err != nil {
    fmt.Println(err)
  }
  // upload test.png
  err = putObject("./testdata/test.png", sourceBucket, "test.png")
  if err != nil {
    fmt.Println(err)
  }
}

func teardown() {
  err := deleteFileObjectFromDestinationBucket("test.png")
  if err != nil {
    fmt.Println(err)
  }
  err = destroyBucket(destBucket)
  if err != nil {
    fmt.Println(err)
  }
  err = destroyBucket(sourceBucket)
  if err != nil {
    fmt.Println(err)
  }
}

func createBucket(b string) error {
  svc := s3.New(sess)
  _, err := svc.CreateBucket(&s3.CreateBucketInput{
    Bucket: aws.String(b),
  })
  if err != nil {
    return err
  }
  return nil
}

func destroyBucket(b string) error {
  svc := s3.New(sess)
  _, err := svc.DeleteBucket(&s3.DeleteBucketInput{
    Bucket: aws.String(b),
  })
  if err != nil {
    return err
  }
  return nil
}

// Delete files from
func deleteFileObjectFromDestinationBucket(obj string) error {
  svc := s3.New(sess)
  _, err := svc.DeleteObjects(&s3.DeleteObjectsInput{
    Bucket: aws.String(destBucket),
    Delete: &s3.Delete{
      Objects: []*s3.ObjectIdentifier{
        { Key: aws.String(obj) },
      },
    },
  })
  if err != nil {
    return err
  }
  return nil
}
