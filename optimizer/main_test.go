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
  // TODO: move buckets to env vars
  sourceBucket string = "image-optimizer-bucket-src-test"
  destBucket string = "image-optimizer-bucket-dest-test"
)

func TestMain(m *testing.M) {
  setup()
  retCode := m.Run()
  // TODO
  // teardown()
  os.Exit(retCode)
}

func TestCreateAndDestroy(t *testing.T) {
  group := &Group {
    Target {
      sourceBucket,
      "",
    },
    Target {
      destBucket,
      "this/is/a/path/",
    },
    []Directive{
      Directive {
        "1000w.jpg",
        []string{"*.jpg", "*.png"},
        "convert {source} -quality 50 -resize x1000> {destination}",
      },
      Directive {
        "800w.jpg",
        []string{"*.jpg", "*.png"},
        "convert {source} -quality 50 -resize x800> {destination}",
      },
      Directive {
        "400w.jpg",
        []string{"*.jpg", "*.png"},
        "convert {source} -quality 50 -resize x400> {destination}",
      },
      Directive {
        "1000w.webp",
        []string{"*.jpg", "*.png"},
        "convert {source} -quality 50 -define webp:lossless=true -resize x1000> {destination}",
      },
      Directive {
        "800w.webp",
        []string{"*.jpg", "*.png"},
        "convert {source} -quality 50 -define webp:lossless=true -resize x800> {destination}",
      },
      Directive {
        "400w.webp",
        []string{"*.jpg", "*.png"},
        "convert {source} -quality 50 -define webp:lossless=true -resize x400> {destination}",
      },
    },
  }
  err := mockCreate("test.png").create(group)
  if err != nil {
    t.Fatal(err)
  }
  err = mockDelete("test.png").destroy(group)
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
            sourceBucket,
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
            sourceBucket,
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

  // Initialize AWS session
  sess = createSession()

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
  err = upload("./testdata/test.png", sourceBucket, "test.png")
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
