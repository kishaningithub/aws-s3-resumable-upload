package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"net/url"
	"os"
	"path"
	"strings"
)

func main() {
	filePath := os.Args[1]
	s3Url := os.Args[2]
	file, err := os.Open(filePath)
	handleErr(err)
	defer file.Close()

	partSize := 10 * 1024 * 1024 // 50 MB
	chunkContent := make([]byte, partSize)
	stat, err := file.Stat()
	size := stat.Size()

	bucket, prefix := parseS3Url(s3Url)
	cfg, err := config.LoadDefaultConfig(context.Background())
	handleErr(err)
	s3Client := s3.NewFromConfig(cfg)

	prefix += "/" + path.Base(file.Name())
	fmt.Printf("Initiating upload...\n")

	upload, err := s3Client.CreateMultipartUpload(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix),
	})
	handleErr(err)
	uploadId := upload.UploadId

	quotient, remainder := divmod(int(size), partSize)
	noOfParts := quotient
	if remainder > 0 {
		noOfParts += 1
	}
	fmt.Printf("Uploading parts... noOfParts=%d\n", noOfParts)
	completedParts := make([]types.CompletedPart, 0, noOfParts)
	for partNumber := 1; partNumber <= noOfParts; partNumber++ {
		chunkContentLength, err := file.Read(chunkContent)
		handleErr(err)
		for {
			fmt.Printf("Starting uploading part partNo=%d\n", partNumber)
			uploadPartResult, err := s3Client.UploadPart(context.Background(), &s3.UploadPartInput{
				Bucket:        aws.String(bucket),
				Key:           aws.String(prefix),
				PartNumber:    int32(partNumber),
				UploadId:      uploadId,
				Body:          bytes.NewReader(chunkContent[:chunkContentLength]),
				ContentLength: int64(chunkContentLength),
			})
			if err != nil {
				fmt.Printf("upload part failed. performing retry... partNo=%d reason=%s\n", partNumber, err)
				continue
			}
			fmt.Printf("Completed uploading part partNo=%d\n", partNumber)
			completedParts = append(completedParts, types.CompletedPart{
				ETag:       uploadPartResult.ETag,
				PartNumber: int32(partNumber),
			})
			break
		}

		handleErr(err)
	}
	_, err = s3Client.CompleteMultipartUpload(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(prefix),
		UploadId: uploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	handleErr(err)
}

func parseS3Url(uri string) (bucket, prefix string) {
	parsedUrl, _ := url.Parse(uri)
	return parsedUrl.Hostname(), strings.Trim(parsedUrl.EscapedPath(), "/")
}

func divmod(numerator, denominator int) (quotient, remainder int) {
	return numerator / denominator, numerator % denominator
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
