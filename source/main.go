package main

import (
	"context"
	//"encoding/json"
	"errors"
	"log"
	"os"
	"math/rand"
	"time"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ericdaugherty/alexa-skills-kit-golang"
)

var a = &alexa.Alexa{ApplicationID: os.Getenv("ALEXA_APPLICATION_ID"), RequestHandler: &HelloWorld{}, IgnoreTimestamp: true}

const cardTitle = "HelloWorld"

// HelloWorld handles reqeusts from the HelloWorld skill.
type HelloWorld struct{}

// Handle processes calls from Lambda
func Handle(ctx context.Context, requestEnv *alexa.RequestEnvelope) (interface{}, error) {
	return a.ProcessRequest(ctx, requestEnv)
}

// OnSessionStarted called when a new session is created.
func (h *HelloWorld) OnSessionStarted(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {

	log.Printf("OnSessionStarted requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	return nil
}

// OnLaunch called with a reqeust is received of type LaunchRequest
func (h *HelloWorld) OnLaunch(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	speechText := "You can ask me about phishing examples or how many stores we have globally"

	log.Printf("OnLaunch requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	response.SetSimpleCard(cardTitle, speechText)
	response.SetOutputText(speechText)
	response.SetRepromptText(speechText)

	response.ShouldSessionEnd = false

	return nil
}

// OnIntent called with a reqeust is received of type IntentRequest
func (h *HelloWorld) OnIntent(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {

	log.Printf("OnIntent requestId=%s, sessionId=%s, intent=%s", request.RequestID, session.SessionID, request.Intent.Name)
	
	switch request.Intent.Name {
	case "GetTalentIntent":
		speechText := "We have over 2200 talents globally"

		response.SetSimpleCard("HelloWorld", speechText)
		response.SetOutputText(speechText)
		response.SetRepromptText(speechText)
	case "GetStoreIntent":
		speechText := "We have 403 stores globally"

		response.SetSimpleCard("HelloWorld", speechText)
		response.SetOutputText(speechText)
		response.SetRepromptText(speechText)
	case "GetPhishingEducationIntent":
		speechText := "Hope we don't meet, but if you need a update your skillset, click on the link below to check the video and quiz on fuse. And I promise it is not a phishing attempt"
		response.SetSimpleCard("HelloWorld", speechText)
		response.SetOutputText(speechText)
		response.SetRepromptText(speechText)
	case "GetPhishingIntent":
		log.Println("phishing intent")

		a := []string{"Sure, we got targeted by teams", "Sure, we got targeted by fake office 365 login page", "Sure, we got targeted by office 365 quarentie", "Sure, we got targeted by fake invoices", "Sure, we got targeted fake roger", "Sure, we got targeted by fake johan", "Sure, we got targeted by fake filip", "Sure, we got targeted by distributors"}
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
		
		speechText := a[1]

		getS3PhishingFile()

		response.SetSimpleCard(cardTitle, speechText)
		response.SetOutputText(speechText)

		log.Printf("Set Output speech, value now: %s", response.OutputSpeech.Text)
	case "AMAZON.HelpIntent":
		log.Println("AMAZON.HelpIntent triggered")
		speechText := "You can ask me about a phishing example"

		response.SetSimpleCard("HelloWorld", speechText)
		response.SetOutputText(speechText)
		response.SetRepromptText(speechText)
	case "AMAZON.FallbackIntent":
		log.Println("AMAZON.FallbackIntent triggered")
		speechText := "You can ask me about a phishing example"

		response.SetSimpleCard("HelloWorld", speechText)
		response.SetOutputText(speechText)
		response.SetRepromptText(speechText)	
	default:
		return errors.New("Invalid Intent")
	}

	return nil
}

// OnSessionEnded called with a reqeust is received of type SessionEndedRequest
func (h *HelloWorld) OnSessionEnded(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {

	log.Printf("OnSessionEnded requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	return nil
}

func main() {
	lambda.Start(Handle)
}

func getS3PhishingFile() error {
	getS3File("wengyretail-dev-www","tv.html")
	return nil
}
//https://lex-web-ui-codebuilddeploy-s0msjg9v0-webappbucket-x3k6d10fuf8i.s3.amazonaws.com/aws-config.js
func getS3File(bucketName, key string) error {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return err
	}
	client := s3.New(cfg)

	log.Printf("s3 GET object %s/%s", bucketName, key)
	result, err := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}).Send()
	if err != nil {
		log.Printf("%s", err)
		return err
	}
	fmt.Printf("result: %s", result)
	return nil  
}

