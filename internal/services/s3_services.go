package services

import (
	"Backend/configs"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	s3Client *s3.Client
	bucket   string
}

func NewAWSService() (*S3Service, error) {
	s3Config := configs.LoadConfig()
	var region = s3Config.AWSRegion
	var bucket = s3Config.S3Bucket
	var accessKey = s3Config.AWSAccessKeyId
	var secretKey = s3Config.AWSSecretAccessKey
	var url = s3Config.S3Endpoint
	var usePathStyle = s3Config.S3UsePathStyle

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           url,
					SigningRegion: region,
				}, nil
			}),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create an Amazon S3 service client access key and so on
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = usePathStyle
	})

	return &S3Service{
		s3Client: s3Client,
		bucket:   bucket,
	}, nil
}

func NewR2Service() (*S3Service, error) {
	s3Config := configs.LoadConfig()
	var bucket = s3Config.S3Bucket
	// Set a default bucket name if it's empty
	if bucket == "" {
		bucket = "pufa-2025"
	}
	var accessKey = s3Config.CloudflareR2AccessId
	var secretKey = s3Config.CloudflareR2AccessKey
	var url = "https://" + s3Config.CloudflareAccountId + ".r2.cloudflarestorage.com/"

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: url,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("apac"),
	)
	if err != nil {
		return nil, err
	}

	// Use path style addressing to avoid bucket name validation issues
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Service{
		s3Client: s3Client,
		bucket:   bucket,
	}, nil
}

func (s *S3Service) UploadFileToAWS(ctx context.Context, directory, key string, file []byte) error {
	key = directory + "/" + key + ".jpg"

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(file),
		ContentType: aws.String("image/jpeg"),
	}

	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Service) UploadFileToR2(ctx context.Context, directory, key string, file []byte, fileType string) error {
	// Determine file extension and content type based on fileType
	var fileExt string
	var contentType string

	switch fileType {
	case "application/pdf":
		fileExt = ".pdf"
		contentType = "application/pdf"
	case "image/jpeg", "image/jpg":
		fileExt = ".jpg"
		contentType = "image/jpeg"
	case "image/png":
		fileExt = ".png"
		contentType = "image/png"
	case "application/zip", "application/x-zip-compressed":
		fileExt = ".zip"
		contentType = "application/zip"
	default:
		// Default to jpg if type is unknown
		fileExt = ".jpg"
		contentType = "image/jpeg"
	}

	// Format the key as directory/key.extension
	key = directory + "/" + key + fileExt

	// Log the bucket and key for debugging
	fmt.Printf("Uploading to R2 - Bucket: %s, Key: %s, Type: %s\n", s.bucket, key, contentType)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(file),
		ContentType: aws.String(contentType),
		// Public read access is handled by bucket policy in Cloudflare R2
	}

	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		fmt.Printf("Error uploading to R2: %v\n", err)
		return err
	}

	fmt.Printf("Successfully uploaded to R2 - Bucket: %s, Key: %s\n", s.bucket, key)
	return nil
}

func (s *S3Service) FileExists(ctx context.Context, directory, slug string) (bool, error) {
	key := directory + "/" + slug + ".jpg"

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.s3Client.HeadObject(ctx, input)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *S3Service) DeleteFile(ctx context.Context, directory, slug string) error {
	key := directory + "/" + slug + ".jpg"
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

// GetFileAWS GetFile GetBucket file from S3
func (s *S3Service) GetFileAWS(directory, slug string) (string, error) {
	key := directory + "/" + slug + ".jpg"
	return "https://id.pufacomputing.live/" + key, nil
}

func (s *S3Service) GetFileR2(directory, slug string) (string, error) {
	// Try to determine if this is a special case for event registrations
	// which could be PDF, JPEG, PNG, or ZIP
	var fileExt string
	if strings.HasPrefix(directory, "event_registrations") {
		// For event registrations, we need to check all possible extensions
		extensions := []string{".pdf", ".jpg", ".png", ".zip"}
		for _, ext := range extensions {
			testKey := directory + "/" + slug + ext
			// Check if file exists with this extension
			input := &s3.HeadObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(testKey),
			}
			_, err := s.s3Client.HeadObject(context.Background(), input)
			if err == nil {
				// File exists with this extension
				fileExt = ext
				break
			}
		}

		// If we couldn't determine the extension, default to .jpg
		if fileExt == "" {
			fileExt = ".jpg"
		}
	} else {
		// For other cases, default to .jpg as before
		fileExt = ".jpg"
	}

	// Format the key as it's stored in R2
	key := directory + "/" + slug + fileExt

	// For Cloudflare R2 with custom domain
	// Add debugging to see what URL is being generated
	fmt.Printf("Generating R2 URL for key: %s\n", key)

	// Add a timestamp to prevent browser caching
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Use the public URL format that works with your Cloudflare R2 setup
	// Add a cache-busting parameter to force browser to reload the image
	url := fmt.Sprintf("https://pufacompsci.my.id/%s?t=%d", key, timestamp)
	fmt.Printf("Generated URL with cache-busting: %s\n", url)
	return url, nil
}
