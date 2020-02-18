package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"fmt"
	"io/ioutil"
 //   "strconv"
    "strings"
	"text/template"
	"bytes"

	"github.com/aws/aws-lambda-go/lambda"
	aws1	"github.com/aws/aws-sdk-go/aws"
	session1 "github.com/aws/aws-sdk-go/aws/session"
	ssm1 "github.com/aws/aws-sdk-go/service/ssm"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ericdaugherty/alexa-skills-kit-golang"
	"database/sql"
	_ "github.com/snowflakedb/gosnowflake"
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

	//var bucketName string
	bucketName := os.Getenv("BUCKET_NAME")

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
			return err
	}
	client := s3.New(cfg)

	log.Printf("s3 GET object %s/%s", bucketName, request.Intent.Name)
	result, err := client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(request.Intent.Name),
	}).Send()
	if err != nil {
			log.Printf("%s", request.Intent.Name)
			return err
	}

	s3objectBytes, err := ioutil.ReadAll(result.Body)
    if err != nil {
        panic(err)
    }
    // create file
	var fileName string
	fileName = "/tmp/" + request.Intent.Name
    f, err := os.Create(fileName)
    defer f.Close()
    if err != nil {
        panic(err)
    }

    bytesWritten, err := f.Write(s3objectBytes)
    if err != nil {
        panic(err)
    }
    //fmt.Printf("Fetched %d bytes for S3Object\n", bytesWritten)
    fmt.Printf("successfully downloaded (%d bytes) data from %s/%s\n to %s", bytesWritten, bucketName, request.Intent.Name, fileName)

    data, err := ioutil.ReadFile(fileName)
    if err != nil {
	  log.Printf("problems reading file")
      fmt.Print(err)
	}

	//time to use the intent

    // define data structure 
    type IntentStruct struct {
		Source string
		Query string
		Say string
		Database string
		Hostname string
		ParameterStore string
		SlotSay string
		SlotQuery string
	  }
	 var intent IntentStruct
	 err = json.Unmarshal(data, &intent)

	  if err != nil {
		  fmt.Println("error: ", err)
	  }
  
	  if intent.Source == "static" {

			fmt.Printf("INFO: %s is a static intent, Will say the following: %s\n", request.Intent.Name, intent.Say);
			response.SetSimpleCard("HelloWorld", intent.Say)
			response.SetOutputText(intent.Say)
			response.SetRepromptText(intent.Say)

		} else if intent.Source == "snowflake" {

			var username string
			var hostname string
			var dbName string

			parameterStore := strings.Split(intent.ParameterStore, "/")
			username = parameterStore[len(parameterStore)-1] //The last value from the parameter is always the username to the DB
			dbparam := "/config/alexa-for-n00bs/"+ intent.Source + "/" + username + "/db" //The structure is db to get the database name from the parameter path
			hostnameparam := "/config/alexa-for-n00bs/"+ intent.Source + "/" + username + "/hostname" //The structure is hostname to get the database name from the parameter path


			//Will create a session, to be used to get the params
			sess, err := session1.NewSessionWithOptions(session1.Options{
				Config:            aws1.Config{Region: aws1.String("eu-west-1")},
				SharedConfigState: session1.SharedConfigEnable,
			})

			if err != nil {
				panic(err)
			}
			ssmsvc := ssm1.New(sess, aws1.NewConfig().WithRegion("eu-west-1"))

			//the username parameter must always be protected by KMS
			withDecryption := true
			param, err := ssmsvc.GetParameter(&ssm1.GetParameterInput{
				Name:           &intent.ParameterStore,
				WithDecryption: &withDecryption,
			})

			if err != nil {
				fmt.Println("error: ", err)
			}
			
			log.Printf("Accessing the following parameters: %s, %s and %s", intent.ParameterStore, dbparam, hostnameparam)
			//log.Printf("ssm med key %s", intent.ParameterStore )

			//log.Printf("doing ssm value:%v",param)


			password := *param.Parameter.Value

			//The db parameter does not need to be encrypted with KMS
			withDecryptionDb := false
			param2, err := ssmsvc.GetParameter(&ssm1.GetParameterInput{
				Name:           &dbparam,
				WithDecryption: &withDecryptionDb,
			})
			if err != nil {
				fmt.Println("error: ", err)
			}
			dbName = *param2.Parameter.Value
			
			//The hostname parameter does not need to be encrypted with KMS
			withDecryptionHostname := false
			param3, err := ssmsvc.GetParameter(&ssm1.GetParameterInput{
				Name:           &hostnameparam,
				WithDecryption: &withDecryptionHostname,
			})
			if err != nil {
				fmt.Println("error: ", err)
			}
			hostname = *param3.Parameter.Value
			//strings.ContainsAny(intent.QueryType,"parameter")

			db, err := sql.Open(intent.Source, username + ":" + password + "@" + hostname + "/" + dbName)
			if err != nil {
				log.Fatal(err)
			}

			defer db.Close()

			if err != nil {
				fmt.Println("Failed to connect", err)
				return err
			}
			rows, err := db.Query(intent.Query)
			if err != nil {
				fmt.Println("Failed to run query", err)
				return err
			}
		
			cols, err := rows.Columns()
			if err != nil {
				fmt.Println("Failed to get columns", err)
				return err
			}


			    // Result is your slice string.
				rawResult := make([][]byte, len(cols))
				result := make([]string, len(cols))
			
				dest := make([]interface{}, len(cols)) // A temporary interface{} slice
				for i, _ := range rawResult {
					dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
				}
	//			fmt.Printf("index %+v value %#vand length %i\n", rawResult,rawResult, len(result))
			
				for rows.Next() {
					err = rows.Scan(dest...)
					if err != nil {
						fmt.Println("Failed to scan row", err)
						return err
					}
			
					for i, raw := range rawResult {
						if raw == nil {
							result[i] = "\\N"
						} else {
							result[i] = string(raw)
						}
					}
			
//					fmt.Printf("%+v and length %i\n", result, len(result))

					//Starting the templating processing with the result from the database
					var tpl bytes.Buffer	
					tmpl, err := template.New("test").Parse(intent.Say)
					if err != nil { panic(err) }
					err = tmpl.Execute(&tpl, result)
					if err != nil { panic(err) }

					fmt.Printf("INFO: %s is a snowflake source, Will say the following: %s\n", request.Intent.Name, tpl.String());
					
					response.SetSimpleCard("Alexa framwork for n00bs github project", tpl.String())
					response.SetOutputText(tpl.String())
					response.SetRepromptText(tpl.String())


		
			}

		} else {
		fmt.Printf("ERROR: Got the following intent that did not match the Souce value (wrong structure or missing fields?): %s with following source: %s \n", request.Intent.Name, intent.Source);
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
