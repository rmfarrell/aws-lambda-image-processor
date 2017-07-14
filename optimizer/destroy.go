package optimizer

import (
  "fmt"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/service/s3"
  // "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Destroy an endpoint in a group's destination
func (ev *S3Event) destroy(group *Group) error {

  // Key parts for deleted key from origin
  kp := newKeyParts(ev.Records[0].S3.Object.Key)

  keys, err := listAllKeysWithPrefx(
    group.Destination.BucketName,
    fmt.Sprintf("%s%s", group.Destination.Root, kp.slug),
  )
  if err != nil {
    return err
  }
  fmt.Println(keys)
  return nil
}


// ------------------------------ Helpers ------------------------------

// Get all the keys from a bucket given a prefix
func listAllKeysWithPrefx(bucket, pfx string) ([]string, error) {
  svc := s3.New(sess)
  out := []string{}
  result, err := svc.ListObjects(&s3.ListObjectsInput{
      Bucket:  aws.String(bucket),
      MaxKeys: aws.Int64(100),
      Prefix:  aws.String(pfx),
  })
  if err != nil {
    return out, err
  }
  for _, c := range result.Contents {
    out = append(out, *c.Key)
  }
  return out, nil
}

// Delete a bunch of keys from a bucket concurrently
func massDeleteKeys(bucket string, keys []string) error {
  return nil
}
