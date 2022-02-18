package file

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"upload-file-s3/internal/config"
)

type Service interface {
	UploadFile(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)
	Download(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)
}

type service struct {
	session *session.Session
	env     config.Env
}

func NewService(session *session.Session, env config.Env) Service {
	return &service{
		session: session,
		env:     env,
	}
}

type UploadResponse struct {
}

func (s service) UploadFile(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Printf("--%#v", req)
	f, filename, err := s.parseMultipartForm(req)
	if err != nil {
		log.Printf("s.parseMultiparForm: %v", err)
		return toJson(http.StatusInternalServerError, UploadResponse{})
	}
	defer f.Close()
	upload := s3manager.NewUploader(s.session)
	_, err = upload.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.env.S3BucketName),
		Key:    aws.String(filename),
		Body:   f,
	})
	if err != nil {
		log.Printf("upload.Upload: %v", err)
		return toJson(http.StatusInternalServerError, UploadResponse{})
	}

	return toJson(http.StatusOK, UploadResponse{})
}

func (s service) Download(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	filename := req.QueryStringParameters["filename"]

	target := fmt.Sprintf("%v/%v", s.env.LocalPath, filename)
	log.Printf("start download from %v", target)
	_, err := os.Stat(target)
	if err == nil {
		_ = os.Remove(target)
	}

	f, err := os.Create(target)
	if err != nil {
		log.Printf("os.Create: %v", err)
		return toJson(http.StatusInternalServerError, UploadResponse{})
	}
	defer f.Close()

	download := s3manager.NewDownloader(s.session)
	_, err = download.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(s.env.S3BucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		log.Printf("download.Download: %v", err)
		return toJson(http.StatusInternalServerError, UploadResponse{})
	}

	fileBytes, err := ioutil.ReadFile(target)
	if err != nil {
		return toJson(http.StatusInternalServerError, UploadResponse{})
	}

	resp := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"content-type":        req.Headers["content-type"],
			"Content-Disposition": fmt.Sprintf("attachment; filename=%v", filename),
		},
		Body:            string(fileBytes),
		IsBase64Encoded: true,
	}
	return &resp, nil
}

func (s service) parseMultipartForm(req events.APIGatewayProxyRequest) (multipart.File, string, error) {
	r := http.Request{}
	r.Header = make(map[string][]string)
	for k, v := range req.Headers {
		r.Header.Set(k, v)
	}

	body, err := base64.StdEncoding.DecodeString(req.Body)
	if err != nil {
		log.Printf("err while decode: %v", err)
		return nil, "", err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	if err != nil {
		log.Printf("err while NopCloser: %v", err)
		return nil, "", err
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Printf("ParseMultiparForm: %v", err)
		return nil, "", err
	}

	file, handler, err := r.FormFile("name")
	if err != nil {
		log.Printf("r.FormFile: %v", err)
		return nil, "", err
	}

	log.Printf("request filename is %v", handler.Filename)
	err = initRootFolder(s.env.LocalPath)
	if err != nil {
		log.Printf("initRootFolder: %v", err)
		return nil, "", err
	}

	target := fmt.Sprintf("%v/%v", s.env.LocalPath, handler.Filename)
	log.Printf("start create %v", target)
	//f, err := os.Create(target)
	//if err != nil {
	//	log.Printf("os.Creat: %v", err)
	//	return nil, "", err
	//}

	//_, err = io.Copy(f, file)
	//if err != nil {
	//	log.Printf("io.Copy: %v", err)
	//	return nil, "", err
	//}
	return file, handler.Filename, nil
}

func initRootFolder(target string) error {
	_, err := os.Stat(target)
	if err != nil {
		err = os.MkdirAll(target, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func toJson(statusCode int, body interface{}) (*events.APIGatewayProxyResponse, error) {
	b, _ := json.Marshal(body)
	resp := events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       string(b),
	}
	return &resp, nil
}
