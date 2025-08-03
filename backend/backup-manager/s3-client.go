package backup_manager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"pg_bckup_mgr/auth"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	ConnectionName string
	EndpointURL    string
	Region         string
	BucketName     string
	AccessKeyID    string
	SecretKeyID    string
	UseSSL         bool
	VerifySSL      bool
	client         *s3.Client
}

func NewS3Client(connectionName, endpointURL, region, bucketName, accessKeyID, secretKeyID string, useSSL, verifySSL bool) (*S3Client, error) {
	decryptedAccessKeyID, _ := auth.DecryptPassword(accessKeyID)
	decryptedSecretKeyID, _ := auth.DecryptPassword(secretKeyID)
	s3Client := &S3Client{
		ConnectionName: connectionName,
		EndpointURL:    endpointURL,
		Region:         region,
		BucketName:     bucketName,
		AccessKeyID:    decryptedAccessKeyID,
		SecretKeyID:    decryptedSecretKeyID,
		UseSSL:         useSSL,
		VerifySSL:      verifySSL,
	}

	if err := s3Client.initializeClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	return s3Client, nil
}

func (s *S3Client) initializeClient() error {
	ctx := context.Background()

	creds := credentials.NewStaticCredentialsProvider(s.AccessKeyID, s.SecretKeyID, "")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s.Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	var client *s3.Client
	if s.EndpointURL != "" {
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s.EndpointURL)
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}

	s.client = client
	return nil
}

func (s *S3Client) CreateBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := &s3.CreateBucketInput{
		Bucket: aws.String(s.BucketName),
	}

	if s.Region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s.Region),
		}
	}

	_, err := s.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", s.BucketName, err)
	}

	return nil
}

func (s *S3Client) DeleteBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	listInput := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.BucketName),
		MaxKeys: aws.Int32(1),
	}

	listOutput, err := s.client.ListObjectsV2(ctx, listInput)
	if err != nil {
		return fmt.Errorf("failed to check if bucket %s is empty: %w", s.BucketName, err)
	}

	if len(listOutput.Contents) > 0 {
		return fmt.Errorf("bucket %s is not empty, cannot delete", s.BucketName)
	}

	_, err = s.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(s.BucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket %s: %w", s.BucketName, err)
	}

	return nil
}

func (s *S3Client) UploadFile(filePath string) error {
	// Timeout updated to 30 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	fileName := filepath.Base(filePath)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %s to bucket %s: %w", filePath, s.BucketName, err)
	}

	return nil
}

func (s *S3Client) ListFiles() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var files []string
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.BucketName),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list files in bucket %s: %w", s.BucketName, err)
		}

		for _, object := range output.Contents {
			if object.Key != nil {
				files = append(files, *object.Key)
			}
		}
	}

	return files, nil
}

func (s *S3Client) DeleteFile(filePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file %s from bucket %s: %w", filePath, s.BucketName, err)
	}

	return nil
}

func (s *S3Client) BucketExists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.BucketName),
	})
	if err != nil {
		var noBucket *types.NotFound
		if errors.As(err, &noBucket) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if bucket %s exists: %w", s.BucketName, err)
	}

	return true, nil
}

func (s *S3Client) DownloadFile(key, localPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer file.Close()

	downloader := manager.NewDownloader(s.client)

	_, err = downloader.Download(ctx, file, &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file %s from bucket %s: %w", key, s.BucketName, err)
	}

	return nil
}

func (s *S3Client) TestConnection() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.BucketName),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false
	}

	return true
}
