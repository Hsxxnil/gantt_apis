package util

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"hta/config"
	s3pkg "hta/internal/interactor/pkg/aws/s3"
	"hta/internal/interactor/pkg/util/log"
	"strings"
)

// UploadToS3 is used to upload file to s3
func UploadToS3(inputBase64, filePath string) (url string, err error) {
	sg := s3pkg.NewAmazonStorage(config.S3BucketName)
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
