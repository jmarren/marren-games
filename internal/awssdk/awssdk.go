package awssdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UploadResult struct {
	Path string `json:"path" xml:"path"`
}

var (
	uploader   *manager.Uploader
	downloader *manager.Downloader
	client     *s3.Client
)

func InitAWS() {
	// AWS SDK
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	uploader = manager.NewUploader(client)
	downloader = manager.NewDownloader(client)

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String("ask-away"),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("first page results:")
	for _, object := range output.Contents {
		log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
	}
}

func UploadToS3(fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file:", err)
		return err
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	key := "test-2.png"
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		return err
	}
	// Read the contents of the file into a buffer
	_, uploadErr := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String("ask-away"),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if uploadErr != nil {
		fmt.Println(uploadErr)
		return uploadErr
	}
	return nil
}
