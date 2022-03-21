# aws-s3-resumable-upload

Resilent file upload utilitity for AWS S3.

Faced the problem(s) where you have several gigabytes of data to upload in S3 but your upload speed is very slow? Then this CLI is for you.

This tool never gives up on your upload(network outages, aws credential timeouts etc) and only exits when the upload is truly complete. This internally uses AWS S3 multi part upload feature, to upload the large file in chunks so that if an upload fails only the failed chunk is retried and not the whole file


## Usage

```bash
aws-s3-resumable-upload ./very-large-file.txt s3://my-bucket/specific-path
```
