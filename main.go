package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jawher/mow.cli"
	"github.com/valyala/fasttemplate"
)

func main() {
	app := cli.App("aws-poc", "Retrieves data from a file found in an S3 bucket")

	accessKeyId := app.String(cli.StringOpt{
		Name:   "aws-access-key-id",
		EnvVar: "AWS_ACCESS_KEY_ID",
		Desc:   "The AWS access key id",
	})

	secretAccessKey := app.String(cli.StringOpt{
		Name:   "aws-secret-access-key",
		EnvVar: "AWS_SECRET_ACCESS_KEY",
		Desc:   "The AWS secret access key",
	})

	region := app.String(cli.StringOpt{
		Name:   "aws-region",
		EnvVar: "AWS_REGION",
		Desc:   "The AWS region",
	})

	bucket := app.String(cli.StringOpt{
		Name:   "aws-bucket",
		EnvVar: "AWS_BUCKET",
		Desc:   "The AWS bucket that stores the file",
	})

	file := app.String(cli.StringOpt{
		Name:   "aws-file",
		EnvVar: "AWS_FILE",
		Desc:   "The file that stores the sensitive information",
	})

	template := app.String(cli.StringOpt{
		Name:   "template",
		EnvVar: "TEMPLATE",
		Desc:   "A text that contains template placeholders that will be replaced with values defined in the 'file' parameter",
	})

	startTag := app.String(cli.StringOpt{
		Name:   "start-tag",
		Value:  "{{",
		EnvVar: "START_TAG",
		Desc:   "The start tag used by the template placeholders",
	})

	endTag := app.String(cli.StringOpt{
		Name:   "end-tag",
		Value:  "}}",
		EnvVar: "END_TAG",
		Desc:   "The end tag used by the template placeholders",
	})

	app.Action = func() {
		printConfig(accessKeyId, secretAccessKey, region, bucket, file, template, startTag, endTag)

		awsSession, err := session.NewSession(&aws.Config{
			Region:      aws.String(*region),
			Credentials: credentials.NewStaticCredentials(*accessKeyId, *secretAccessKey, ""),
		})
		if err != nil {
			log.Fatal(err)
		}

		s3Client := s3.New(awsSession)
		resp, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(*bucket),
			Key:    aws.String(*file),
		})
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		var fileContent map[string]interface{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		if err = json.Unmarshal(body, &fileContent); err != nil {
			log.Fatal(err)
		}

		t := fasttemplate.New(*template, *startTag, *endTag)
		s := t.ExecuteString(fileContent)
		log.Printf("Replaced template:\n%s", s)
	}

	app.Run(os.Args)

}

func printConfig(accessKeyId, secretAccessKey, region, bucket, file, template, startTag, endTag *string) {
	log.Print("Starting app configuration:")
	log.Printf("aws-access-key-id: %s", *accessKeyId)
	log.Printf("aws-secret-access-key: %s", *secretAccessKey)
	log.Printf("aws-region: %s", *region)
	log.Printf("aws-bucket: %s", *bucket)
	log.Printf("aws-file: %s", *file)
	log.Printf("template:\n%s", *template)
	log.Printf("start-tag: %s", *startTag)
	log.Printf("end-tag: %s", *endTag)
}
