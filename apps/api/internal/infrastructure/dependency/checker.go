package dependency

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/haseg/anifusion-canvas/apps/api/internal/config"
)

type Status string

const (
	StatusOK      Status = "ok"
	StatusSkipped Status = "skipped"
	StatusError   Status = "error"
)

type CheckResult struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type Checker struct {
	cfg    config.Config
	client *http.Client
}

func NewChecker(cfg config.Config) *Checker {
	return &Checker{
		cfg: cfg,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (c *Checker) CheckAll(ctx context.Context) []CheckResult {
	return []CheckResult{
		c.CheckDatabase(ctx),
		c.CheckReplicate(ctx),
		c.CheckR2(ctx),
		c.CheckFFmpeg(),
	}
}

func (c *Checker) CheckDatabase(ctx context.Context) CheckResult {
	if c.cfg.DatabaseURL == "" {
		return skipped("database", "DATABASE_URL is not set")
	}

	db, err := sql.Open("mysql", c.cfg.DatabaseURL)
	if err != nil {
		return failed("database", err)
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return failed("database", err)
	}
	if c.cfg.StudioStore == "database" {
		if err := checkDatabaseSchema(pingCtx, db); err != nil {
			return failed("database", err)
		}
	}

	return ok("database", "TiDB/MySQL connection is reachable")
}

func checkDatabaseSchema(ctx context.Context, db *sql.DB) error {
	required := map[string][]string{
		"studio_projects": {"id", "name", "created_at", "updated_at"},
		"studio_frames": {
			"id",
			"project_id",
			"frame_index",
			"image_url",
			"thumbnail_url",
			"kind",
			"note",
			"created_at",
			"updated_at",
		},
		"studio_jobs": {
			"id",
			"project_id",
			"job_type",
			"status",
			"progress",
			"message",
			"result_json",
			"error",
			"version",
			"created_at",
			"updated_at",
		},
	}

	var databaseName string
	if err := db.QueryRowContext(ctx, `SELECT DATABASE()`).Scan(&databaseName); err != nil {
		return err
	}
	for table, columns := range required {
		for _, column := range columns {
			var count int
			err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM information_schema.columns
WHERE table_schema = ? AND table_name = ? AND column_name = ?`, databaseName, table, column).Scan(&count)
			if err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("database schema mismatch: missing %s.%s; run migrations before STUDIO_STORE=database", table, column)
			}
		}
	}

	return nil
}

func (c *Checker) CheckReplicate(ctx context.Context) CheckResult {
	if c.cfg.ReplicateMode != "replicate" {
		return skipped("replicate", "REPLICATE_MODE is demo; Replicate inference is disabled")
	}
	if c.cfg.ReplicateAPIToken == "" {
		return skipped("replicate", "REPLICATE_API_TOKEN is not set")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.replicate.com/v1/account", nil)
	if err != nil {
		return failed("replicate", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.ReplicateAPIToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return failed("replicate", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return CheckResult{
			Name:    "replicate",
			Status:  StatusError,
			Message: fmt.Sprintf("Replicate account check failed with status %d: %v", resp.StatusCode, body),
		}
	}

	return ok("replicate", "Replicate API token is valid")
}

func (c *Checker) CheckR2(ctx context.Context) CheckResult {
	missing := make([]string, 0)
	required := map[string]string{
		"R2_ACCOUNT_ID":        c.cfg.R2AccountID,
		"R2_ACCESS_KEY_ID":     c.cfg.R2AccessKeyID,
		"R2_SECRET_ACCESS_KEY": c.cfg.R2SecretAccessKey,
		"R2_BUCKET":            c.cfg.R2Bucket,
		"R2_ENDPOINT_URL":      c.cfg.R2EndpointURL,
	}

	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return skipped("r2", "missing config: "+fmt.Sprint(missing))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(c.cfg.R2Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			c.cfg.R2AccessKeyID,
			c.cfg.R2SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return failed("r2", err)
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(c.cfg.R2EndpointURL)
		options.UsePathStyle = true
	})

	checkCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	_, err = client.HeadBucket(checkCtx, &s3.HeadBucketInput{
		Bucket: aws.String(c.cfg.R2Bucket),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return failed("r2", fmt.Errorf("bucket not found: %s", c.cfg.R2Bucket))
		}
		return failed("r2", err)
	}

	return ok("r2", "R2 bucket is reachable")
}

func (c *Checker) CheckFFmpeg() CheckResult {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return failed("ffmpeg", err)
	}

	return ok("ffmpeg", "ffmpeg found at "+path)
}

func ok(name string, message string) CheckResult {
	return CheckResult{Name: name, Status: StatusOK, Message: message}
}

func skipped(name string, message string) CheckResult {
	return CheckResult{Name: name, Status: StatusSkipped, Message: message}
}

func failed(name string, err error) CheckResult {
	return CheckResult{Name: name, Status: StatusError, Message: err.Error()}
}
