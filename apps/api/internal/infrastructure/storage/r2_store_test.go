package storage

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/haseg/anifusion-canvas/apps/api/internal/config"
)

func TestDecodeDataURL(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		contentType string
		data        string
	}{
		{
			name:        "base64 image",
			value:       "data:image/png;base64,aGVsbG8=",
			contentType: "image/png",
			data:        "hello",
		},
		{
			name:        "url encoded text",
			value:       "data:text/plain,hello%20world",
			contentType: "text/plain",
			data:        "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := DecodeDataURL(tt.value)
			if err != nil {
				t.Fatalf("decode data URL failed: %v", err)
			}
			if decoded.ContentType != tt.contentType {
				t.Fatalf("expected content type %q, got %q", tt.contentType, decoded.ContentType)
			}
			if string(decoded.Data) != tt.data {
				t.Fatalf("expected data %q, got %q", tt.data, string(decoded.Data))
			}
		})
	}
}

func TestDecodeDataURLRejectsInvalidValues(t *testing.T) {
	tests := []string{
		"https://example.test/image.png",
		"data:image/png;base64,***",
		"data:image/png;base64",
	}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			if _, err := DecodeDataURL(value); err == nil {
				t.Fatalf("expected decode to fail")
			}
		})
	}
}

func TestR2StorePutDataURL(t *testing.T) {
	client := newFakeS3Client()
	store := NewR2StoreWithClient(client, "studio-bucket", "https://assets.example.test")

	object, err := store.PutDataURL(context.Background(), "projects/demo/frames/frame 1.png", "data:image/png;base64,aGVsbG8=")
	if err != nil {
		t.Fatalf("put data URL failed: %v", err)
	}

	if object.Key != "projects/demo/frames/frame 1.png" {
		t.Fatalf("unexpected key: %q", object.Key)
	}
	if object.URL != "https://assets.example.test/projects/demo/frames/frame%201.png" {
		t.Fatalf("unexpected public URL: %q", object.URL)
	}
	if object.ContentType != "image/png" || object.Size != 5 {
		t.Fatalf("unexpected object metadata: %#v", object)
	}
	if client.putBucket != "studio-bucket" || client.putKey != object.Key {
		t.Fatalf("unexpected put target bucket=%q key=%q", client.putBucket, client.putKey)
	}
	if string(client.putBody) != "hello" {
		t.Fatalf("unexpected uploaded body: %q", string(client.putBody))
	}
}

func TestNewR2StoreRequiresConfig(t *testing.T) {
	if _, err := NewR2Store(context.Background(), config.Config{}); err == nil {
		t.Fatalf("expected missing config to fail")
	}
}

func TestR2StoreObjectURIWithoutPublicBaseURL(t *testing.T) {
	client := newFakeS3Client()
	store := NewR2StoreWithClient(client, "studio-bucket", "")

	object, err := store.PutBytes(context.Background(), "projects/demo/input.png", "image/png", []byte("hello"))
	if err != nil {
		t.Fatalf("put bytes failed: %v", err)
	}
	if object.URL != "r2://studio-bucket/projects/demo/input.png" {
		t.Fatalf("unexpected object URI: %q", object.URL)
	}
}

func TestR2StoreGetAndDeleteObject(t *testing.T) {
	client := newFakeS3Client()
	client.getBody = []byte("image")
	client.getContentType = "image/png"
	store := NewR2StoreWithClient(client, "studio-bucket", "")

	body, contentType, err := store.GetObject(context.Background(), "projects/demo/input.png")
	if err != nil {
		t.Fatalf("get object failed: %v", err)
	}
	if string(body) != "image" || contentType != "image/png" {
		t.Fatalf("unexpected get result body=%q contentType=%q", string(body), contentType)
	}
	if client.getBucket != "studio-bucket" || client.getKey != "projects/demo/input.png" {
		t.Fatalf("unexpected get target bucket=%q key=%q", client.getBucket, client.getKey)
	}

	if err := store.DeleteObject(context.Background(), "projects/demo/input.png"); err != nil {
		t.Fatalf("delete object failed: %v", err)
	}
	if client.deleteKey != "projects/demo/input.png" {
		t.Fatalf("unexpected delete key: %q", client.deleteKey)
	}
}

type fakeS3Client struct {
	putBucket      string
	putKey         string
	putBody        []byte
	putContentType string
	getBody        []byte
	getContentType string
	getBucket      string
	getKey         string
	deleteKey      string
}

func newFakeS3Client() *fakeS3Client {
	return &fakeS3Client{}
}

func (c *fakeS3Client) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	c.putBucket = aws.ToString(input.Bucket)
	c.putKey = aws.ToString(input.Key)
	c.putContentType = aws.ToString(input.ContentType)
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	c.putBody = body
	return &s3.PutObjectOutput{}, nil
}

func (c *fakeS3Client) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	c.getBucket = aws.ToString(input.Bucket)
	c.getKey = aws.ToString(input.Key)
	return &s3.GetObjectOutput{
		Body:        io.NopCloser(bytes.NewReader(c.getBody)),
		ContentType: aws.String(c.getContentType),
	}, nil
}

func (c *fakeS3Client) DeleteObject(_ context.Context, input *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	c.deleteKey = aws.ToString(input.Key)
	return &s3.DeleteObjectOutput{}, nil
}
