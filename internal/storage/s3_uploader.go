package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"

	// "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Uploader struct {
	Bucket string
	Prefix string
	Client *s3.Client
}

func NewS3Uploader(bucket, prefix string, client *s3.Client) *S3Uploader {
	prefix = strings.Trim(prefix, "/")
	return &S3Uploader{
		Bucket: bucket,
		Prefix: prefix,
		Client: client,
	}
}

// UploadDir uploads all files under localDir to s3://bucket/prefix/{jobID}/...
func (u *S3Uploader) UploadDir(ctx context.Context, jobID string, localDir string) error {
	up := manager.NewUploader(u.Client)

	return filepath.WalkDir(localDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		key := u.keyFor(jobID, rel)

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		contentType := detectContentType(path, f)

		_, err = up.Upload(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(u.Bucket),
			Key:         aws.String(key),
			Body:        f,
			ContentType: aws.String(contentType),
		})
		return err
	})
}

/* func (u *S3Uploader) UploadDir(ctx context.Context, jobID, localDir string) error {
	tmClient := transfermanager.New(u.Client)

	// Build the upload input
	input := &transfermanager.UploadDirectoryInput{
		Bucket:    aws.String(u.Bucket),
		Source:    aws.String(localDir),
		KeyPrefix: aws.String(fmt.Sprintf("%s/%s", u.Prefix, jobID)),
	}

	// Call UploadDirectory
	output, err := tmClient.UploadDirectory(ctx, input)
	if err != nil {
		return fmt.Errorf("upload directory failed: %w", err)
	}

	fmt.Printf("Uploaded %d object, %d failures\n", output.ObjectsUploaded, output.ObjectsFailed)

	return nil
} */

func (u *S3Uploader) keyFor(jobID string, relPath string) string {
	relPath = filepath.ToSlash(relPath)
	if u.Prefix == "" {
		return fmt.Sprintf("%s/%s", jobID, relPath)
	}
	return fmt.Sprintf("%s/%s/%s", u.Prefix, jobID, relPath)
}

func detectContentType(path string, r io.ReadSeeker) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			// .m3u8, .ts may not always be mapped on every system; mime helps when available.
			return ct
		}
	}

	// Fallback: default binary
	_, _ = r.Seek(0, 0)
	return "application/octet-stream"
}

// func detectContentType(path string) string {
// 	ext := strings.ToLower(filepath.Ext(path))

// 	switch ext {
// 	case ".m3u8":
// 		return "application/vnd.apple.mpegurl"
// 	case ".ts":
// 		return "video/mp2t"
// 	}

// 	if ct := mime.TypeByExtension(ext); ct != "" {
// 		return ct
// 	}

// 	return "application/octet-stream"
// }
