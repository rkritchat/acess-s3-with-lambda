package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"log"
	"os"
)

type Cfg struct {
	AwsSession *session.Session
	Env        Env
}

type Env struct {
	LocalPath    string
	S3BucketName string
}

func InitConfig() Cfg {
	env := initEnv()
	return Cfg{
		AwsSession: initAwsSession(),
		Env:        env,
	}
}

func initEnv() Env {
	return Env{
		LocalPath:    os.Getenv("LOCAL_PATH"),
		S3BucketName: os.Getenv("S3_BUCKET_NAME"),
	}
}

func initAwsSession() *session.Session {
	s, err := session.NewSession(
		&aws.Config{
			Region: aws.String(os.Getenv("AWS_REGION")),
		})
	if err != nil {
		log.Fatalln(err)
	}
	return s
}
