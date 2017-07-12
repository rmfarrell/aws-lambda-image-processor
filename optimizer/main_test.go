package optimizer

import (
  "testing"
  "os"
)

func TestMain(m *testing.M) {
  setup()
  retCode := m.Run()
  teardown()
  os.Exit(retCode)
}

func TestHandleS3Event(t *testing.T) {}


// ---------------------- Helpers ----------------------

// Create a mock s3 create event
func mockCreate() S3Event {
  return S3Event {
    []Record {
      Record {
        "ObjectCreated:Put",
        S3 {
          Bucket {
            "carrot-image-handler-test",
            "arn:aws:s3:::carrot-image-handler-test",
          },
          Object {
            "test.png",
            0,
          },
        },
      },
    },
  }
}

// Create a mock s3 delete event
func mockDelete() S3Event {
  return S3Event {
    []Record {
      Record {
        "ObjectRemoved:Delete",
        S3 {
          Bucket {
            "carrot-image-handler-test",
            "arn:aws:s3:::carrot-image-handler-test",
          },
          Object {
            "test.png",
            0,
          },
        },
      },
    },
  }
}


// ---------------------- Setup/teardown ----------------------

// TODO
func setup() {}

// TODO
func teardown() {}
