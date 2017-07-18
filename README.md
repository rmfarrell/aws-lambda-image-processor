## AWS S3-based Lambda for handling transformation of media.
Process all your images and media in real time.

This lambda is designed to listen to events on an AWS S3 bucket, process files
effected via the command line actions and then re-upload them to one or more destination buckets.

# Setup and deployment
1. Install Apex
2. Create a source bucket
3. Configure a source bucket to [activate your lambda on events] (http://docs.aws.amazon.com/AmazonS3/latest/UG/SettingBucketNotifications.html)\
4. Create at least one destination bucket
5. Create a role for your lambda that has at, minimum, read/write access to your
source bucket and all destination buckets. [more info](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create.html)
6. Clone this repo, `cd` into the directory
7. run `cp ./project.json.sample project.json` and update the configuration with
the appropriate values. Replace the `role` attribute with the IAM role you created above.
8. run `apex deploy`
9. Modify the configuration along the lines outlined below. You can use
`config.yml` as an example configuration

# Configuration
Configure the behavior of the lambda via any number of `.yml` files that live in
the source bucket. `groups.yml` & `directives.yml` are a useful template for this.
Just change some values around and drop it in the root of the source bucket.

Any number of `.yml` files can be used. At runtime they will be combined into a single
configuration object.

## Groups
Groups are for grouping together files that go to the same destination bucket. An
example configuration might look like this:

### Single origin/destination bucket
```
groups:
  - destination: a-bucket-name-that-you-own
    directives: [a-directive, another-directive]
```
### Multiple destinations
```
groups:
  - root: a-root-directory/
    destination: a-bucket-name-that-you-own
    directives: [a-directive, another-directive]
  - root: another-root-directory/
    destination: another-bucket
    directives: [a-directive, another-directive]
```
The `root` property, in this latter case will apply the directives each file uploaded
to `a-root-directory` directory in the source bucket and upload the finished file to the
`destination` bucket.

## Directives
Directives take 4 properties: `endpoint`, `matcher`, `command`, and `acl`.
- `endpoint` is the s3 endpoint the processed file is targeting (e.g., `//s3.amazonaws.com/my-destination-bucket/test.jpg`)
- `matcher` is the glob pattern used to determine whether to apply the command
- `command` is the command itself.
- `acl` takes a "canned" AWS ACL setting (default: `private`, you probably want either
  `private`, `public-read` or [see more options](http://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html))
The directive automatically has access to three variables:
- `filename`: the name of the original file.
- `basename`: the basename of the original file.
- `extension`: the extension of the original file.

They can be grouped for convenience and clarity. As shown above they are referenced
by the groups' directive property

### Examples
#### Imagemagick processing of an image
- Resize e.g., `test.jpg` to 1000px if the original size is larger.
- Add progressive downloading
- upload to the destination bucket with the key `test/1000w.jpg`
```
resize:
  - endpoint: '%{basename}/1000w.jpg'
    matcher: ['*.jpg', '*.png']
    command: convert %{filename} -quality 50 -interlace Plane -resize x1000> 1000w.jpg
    acl: public-read
```
#### ffmpeg conversion of .gif
- convert gifs to mp4
- upload both the gif and mp4 to `{basename}.gif` and `{basename}.mp4`
```
convert_gif_to_mp4:
  - endpoint: '%{basename}.mp4'
    matcher: ['*.gif']
    command: ffmpeg -i %{filename} %{endpoint}
    acl: public-read
  - endpoint: '%{basename}.gif'
    matcher: ['*.gif']
    acl: public-read
```

# Questions

### 1. Why this?
The combination of S3 events and a lambda configured to listen to those events is
a powerful tool for organizing files, especially media. Sure, there are other such
lambdas and plenty of great services that do essentially the same thing, but what
sets this one apart is:
1. One source bucket can supply multiple destination buckets. No need to create
two buckets for every project.
2. Media transformation use regular ol' command line. This allows for greater control
over your transformations than some other libraries.
3. Configure your project in the source bucket itself. The config files will be read by
the lambda after initialization so there is no need redeploy the lambda for every
change to the configuration.

### 2. Can my origin bucket have more than one destination bucket?
Damn yeah, it can! I don't like resourcing AWS infrastructure any more than the
next fellow or gal. That's why this lambda is designed to easily extend it's
functionality to multiple destinations with very little additional configuration.

Just add another folder in your S3 and specify a new `group` in your `.yml` file
whose `root` points to that directory, and create a new destination bucket, making
sure your lambda has the appropriate permissions for both

### 3. Can I configure this for use a single bucket?
Yes, but it is highly not recommended. You could configure this so that your
source bucket listens for events on that buckets and then also makes changes to it,
but that would in turn trigger a new event, and then another. Where will it end?
In tears; that's where.

### 4. Can I still safely make changes to my destination bucket?
Yep.

### 5. Where are my logs?
AWS CloudWatch. [More info](//apex.run/#viewing-log-output)
