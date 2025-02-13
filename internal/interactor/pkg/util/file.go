package util

import (
	"encoding/base64"
	s3pkg "gantt/internal/interactor/pkg/aws/s3"
	"gantt/internal/interactor/pkg/util/log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadToS3 is used to upload file to s3
func UploadToS3(inputBase64, filePath string) (url string, err error) {
	s3bucketName := "pm-s3-bucket"
	sg := s3pkg.NewAmazonStorage(s3bucketName)
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(inputBase64))
	object := &s3.PutObjectInput{
		Key:  aws.String(filePath),
		Body: reader,
	}

	info, err := sg.Upload(object)
	if err != nil {
		log.Error(err)
		return "", err
	}

	url = info.Location
	return url, nil
}
