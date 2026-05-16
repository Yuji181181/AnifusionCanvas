package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/haseg/anifusion-canvas/apps/api/internal/config"
)

type s3Client interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type R2Store struct {
	client        s3Client
	bucket        string
	publicBaseURL string
}

type StoredObject struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

func NewR2Store(ctx context.Context, cfg config.Config) (*R2Store, error) {
	if strings.TrimSpace(cfg.R2Bucket) == "" {
		return nil, fmt.Errorf("R2_BUCKET is required")
	}
	if strings.TrimSpace(cfg.R2EndpointURL) == "" {
		return nil, fmt.Errorf("R2_ENDPOINT_URL is required")
	}
	if strings.TrimSpace(cfg.R2AccessKeyID) == "" {
		return nil, fmt.Errorf("R2_ACCESS_KEY_ID is required")
	}
	if strings.TrimSpace(cfg.R2SecretAccessKey) == "" {
		return nil, fmt.Errorf("R2_SECRET_ACCESS_KEY is required")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.R2Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2AccessKeyID,
			cfg.R2SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.R2EndpointURL)
		options.UsePathStyle = true
	})

	return NewR2StoreWithClient(client, cfg.R2Bucket, cfg.R2PublicBaseURL), nil
}

func NewR2StoreWithClient(client s3Client, bucket string, publicBaseURL string) *R2Store {
	return &R2Store{
		client:        client,
		bucket:        bucket,
		publicBaseURL: publicBaseURL,
	}
}

func (s *R2Store) PutDataURL(ctx context.Context, key string, dataURL string) (StoredObject, error) {
	decoded, err := DecodeDataURL(dataURL)
	if err != nil {
		return StoredObject{}, err
	}

	return s.PutBytes(ctx, key, decoded.ContentType, decoded.Data)
}

func (s *R2Store) PutBytes(ctx context.Context, key string, contentType string, data []byte) (StoredObject, error) {
	if strings.TrimSpace(key) == "" {
		return StoredObject{}, fmt.Errorf("object key is required")
	}
	if strings.TrimSpace(s.bucket) == "" {
		return StoredObject{}, fmt.Errorf("r2 bucket is required")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return StoredObject{}, err
	}

	return StoredObject{
		Key:         key,
		URL:         s.objectURL(key),
		ContentType: contentType,
		Size:        int64(len(data)),
	}, nil
}

func (s *R2Store) GetObject(ctx context.Context, key string) ([]byte, string, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", err
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}
	return body, contentType, nil
}

func (s *R2Store) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

type DecodedDataURL struct {
	ContentType string
	Data        []byte
}

func DecodeDataURL(value string) (DecodedDataURL, error) {
	if !strings.HasPrefix(value, "data:") {
		return DecodedDataURL{}, fmt.Errorf("data URL must start with data:")
	}

	metadata, payload, ok := strings.Cut(strings.TrimPrefix(value, "data:"), ",")
	if !ok {
		return DecodedDataURL{}, fmt.Errorf("data URL payload is missing")
	}

	parts := strings.Split(metadata, ";")
	contentType := parts[0]
	if contentType == "" {
		contentType = "text/plain;charset=US-ASCII"
	}
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		return DecodedDataURL{}, fmt.Errorf("invalid data URL media type: %w", err)
	}

	isBase64 := false
	for _, part := range parts[1:] {
		if strings.EqualFold(part, "base64") {
			isBase64 = true
			break
		}
	}

	if isBase64 {
		data, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return DecodedDataURL{}, fmt.Errorf("decode base64 data URL: %w", err)
		}
		return DecodedDataURL{ContentType: contentType, Data: data}, nil
	}

	data, err := url.PathUnescape(payload)
	if err != nil {
		return DecodedDataURL{}, fmt.Errorf("decode data URL payload: %w", err)
	}
	return DecodedDataURL{ContentType: contentType, Data: []byte(data)}, nil
}

func (s *R2Store) objectURL(key string) string {
	if s.publicBaseURL == "" {
		return "r2://" + s.bucket + "/" + key
	}
	return strings.TrimRight(s.publicBaseURL, "/") + "/" + escapeObjectKey(key)
}

func escapeObjectKey(key string) string {
	segments := strings.Split(key, "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
