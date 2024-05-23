package aws

import (
	"foxy/internal/db"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"time"
)

func NewS3Client(config db.SourceConfig) *s3.S3 {
	creds := credentials.NewStaticCredentials(
		*config.Key,
		*config.Secret,
		"",
	)

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      config.Region,
		Credentials: creds,
	}))

	return s3.New(sess)
}

func GetSignedUrl(config db.SourceConfig, key string, expiresIn time.Duration) (string, error) {
	s3Client := NewS3Client(config)
	r, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: config.Bucket,
		Key:    aws.String(key),
	})

	url, err := r.Presign(expiresIn)

	return url, err
}
