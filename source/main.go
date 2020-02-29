package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"fmt"
	"io/ioutil"
    "strconv"
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
type IntentStruct struct {
	Source string
	Query string
	Say string
	Database string
	Hostname string
	ParameterStore string
	SlotName string
  }

var intent IntentStruct

var Debug,err =strconv.ParseBool(os.Getenv("DEBUG_OUTPUT"))


var a = &alexa.Alexa{ApplicationID: os.Getenv("ALEXA_APPLICATION_ID"), RequestHandler: &RequestSkill{}, IgnoreTimestamp: true}

//const cardTitle = "Alexa framwork for n00bs github project"

// I have no clue why this one is needed but did not manage to get rid of it from the methods
type RequestSkill struct{}

// Handle processes calls from Lambda
func Handle(ctx context.Context, requestEnv *alexa.RequestEnvelope) (interface{}, error) {
	return a.ProcessRequest(ctx, requestEnv)
}

// OnSessionStarted called when a new session is created.
func (h *RequestSkill) OnSessionStarted(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {

	if Debug == true {
		log.Printf("OnSessionStarted requestId=%s, sessionId=%s intentName=%s", request.RequestID, session.SessionID, request.Intent.Name)
	}

	return nil
}

// OnLaunch called with a reqeust is received of type LaunchRequest
func (h *RequestSkill) OnLaunch(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	if Debug == true {
		log.Printf("OnLaunch requestId=%s, sessionId=%s, intentName=%s", request.RequestID, session.SessionID,request.Intent.Name)
	}

	//Need to hardcode the intent name to onlaunch as it is actually not a intent
	err := getIntent("OnLaunch" ,&intent)
	if err != nil {
	  log.Printf("Got some problems to fetch the intent file from S3, does it exist?")
	  panic(err)
  }  
  
  err =  executeIntent("OnLaunch", &intent, request, response)
  if err != nil {
	  panic(err)
  }  
  response.ShouldSessionEnd = false
  return nil

}

// OnIntent called with a reqeust is received of type IntentRequest
func (h *RequestSkill) OnIntent(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	if Debug == true {
		log.Printf("OnIntent requestId=%s, sessionId=%s, intent=%s intentslot=%+v", request.RequestID, session.SessionID, request.Intent.Name, request.Intent.Slots)
	}
    err := getIntent(request.Intent.Name ,&intent)
	  if err != nil {
		log.Printf("Got some problems to fetch the intent file from S3, does it exist?")
		panic(err)
	}  
	
	err =  executeIntent(request.Intent.Name, &intent, request, response)
	if err != nil {
		panic(err)
	}  
	return nil
}

// OnSessionEnded called with a reqeust is received of type SessionEndedRequest
func (h *RequestSkill) OnSessionEnded(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {

	log.Printf("OnSessionEnded requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	return nil
}

func main() {
	lambda.Start(Handle)
}

//This code bellow to the getIntent function comes from https://gist.github.com/SQLServerIO/91e63f29c5f13b0f3fc269c2e068a2b5
type mapStringScan struct {
	// cp are the column pointers
	cp []interface{}
	// row contains the final result
	row      map[string]string
	colCount int
	colNames []string
}

func NewMapStringScan(columnNames []string) *mapStringScan {
	lenCN := len(columnNames)
	s := &mapStringScan{
		cp:       make([]interface{}, lenCN),
		row:      make(map[string]string, lenCN),
		colCount: lenCN,
		colNames: columnNames,
	}
	for i := 0; i < lenCN; i++ {
		s.cp[i] = new(sql.RawBytes)
	}
	return s
}

func (s *mapStringScan) Update(rows *sql.Rows) error {
	err := rows.Scan(s.cp...)
	if err != nil {
		return err
	}

	for i := 0; i < s.colCount; i++ {
		if rb, ok := s.cp[i].(*sql.RawBytes); ok {
			s.row[s.colNames[i]] = string(*rb)
			*rb = nil // reset pointer to discard current value to avoid a bug
		} else {
			return fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, s.colNames[i])
		}
	}
	return nil
}

func (s *mapStringScan) Get() map[string]string {
	return s.row
}

/**
  using a string slice
*/
type stringStringScan struct {
	// cp are the column pointers
	cp []interface{}
	// row contains the final result
	row      []string
	colCount int
	colNames []string
}

func NewStringStringScan(columnNames []string) *stringStringScan {
	lenCN := len(columnNames)
	s := &stringStringScan{
		cp:       make([]interface{}, lenCN),
		row:      make([]string, lenCN*2),
		colCount: lenCN,
		colNames: columnNames,
	}
	j := 0
	for i := 0; i < lenCN; i++ {
		s.cp[i] = new(sql.RawBytes)
		s.row[j] = s.colNames[i]
		j = j + 2
	}
	return s
}

func (s *stringStringScan) Update(rows *sql.Rows) error {
	err := rows.Scan(s.cp...)
	if err != nil {
		return err
	}
	j := 0
	for i := 0; i < s.colCount; i++ {
		if rb, ok := s.cp[i].(*sql.RawBytes); ok {
			s.row[j+1] = string(*rb)
			*rb = nil // reset pointer to discard current value to avoid a bug
		} else {
			return fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, s.colNames[i])
		}
		j = j + 2
	}
	return nil
}

func (s *stringStringScan) Get() []string {
	return s.row
}

// rowMapString was the first implementation but it creates for each row a new
// map and pointers and is considered as slow. see benchmark
func rowMapString(columnNames []string, rows *sql.Rows) (map[string]string, error) {
	lenCN := len(columnNames)
	ret := make(map[string]string, lenCN)

	columnPointers := make([]interface{}, lenCN)
	for i := 0; i < lenCN; i++ {
		columnPointers[i] = new(sql.RawBytes)
	}

	err := rows.Scan(columnPointers...)
	if err != nil {
		return nil, err
	}

	for i := 0; i < lenCN; i++ {
		if rb, ok := columnPointers[i].(*sql.RawBytes); ok {
			ret[columnNames[i]] = string(*rb)
		} else {
			return nil, fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, columnNames[i])
		}
	}

	return ret, nil
}


//This function will check if the json intent file exists on S3 and fetch it
func getIntent(intentName string, intent *IntentStruct) (err error) {

	//var bucketName string
	bucketName := os.Getenv("BUCKET_NAME")

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
			log.Printf("Got an error when loading and initating aws config")
			return err
	}
	client := s3.New(cfg)
	if Debug == true {
		log.Printf("Will fetch the intent %s from s3 bucket %s/%s (=bucket/file)", intentName, bucketName, intentName)
	}
	result, err := client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(intentName),
	}).Send()
	if err != nil {
		log.Printf("Got an error when fetching the intent file from S3, does it really exists? Please check manually")
		return err
}
	s3objectBytes, err := ioutil.ReadAll(result.Body)
    if err != nil {
		log.Printf("Got an error when reading the file from S3")
		return err
    }
    // create file
	var fileName string
	fileName = "/tmp/" + intentName
    f, err := os.Create(fileName)
    defer f.Close()
    if err != nil {
		log.Printf("Got an error when trying to create the file (%s) locally from S3", fileName)
		return err
    }

    bytesWritten, err := f.Write(s3objectBytes)
    if err != nil {
		log.Printf("Got an error when writing the file %s", fileName)
		return err
    }
    //fmt.Printf("Fetched %d bytes for S3Object\n", bytesWritten)
	if Debug == true {
		log.Printf("successfully downloaded (%d bytes) data from %s/%s\n to %s", bytesWritten, bucketName, intentName, fileName)
	}
    data, err := ioutil.ReadFile(fileName)
    if err != nil {
	 
		log.Printf("Got an error when trying to read the file (%s)", fileName)
		return err
	}

	if Debug == true {
		log.Printf("Read the file (%s) and will now parse it as a json",fileName)
	}

	 err = json.Unmarshal(data, &intent)

	  if err != nil {
		log.Printf("Got an error when trying to unmarshal the json file to the struct, is there any crap in the intent file? Is it a json file?")
		return err
	  }
	  return nil
}


//This function will parse the intent file and do the magic
func executeIntent(intentName string, intent *IntentStruct, request *alexa.Request,response *alexa.Response) (err error) {	

	if intent.Source == "static" {

		if Debug == true {
			log.Printf("INFO: %s is a static intent, Will say the following: %s\n", request.Intent.Name, intent.Say)
		}
			
		response.SetSimpleCard("Alexa framwork for n00bs github project", intent.Say)
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


		//Will create a session for aws, to be used to get the params
		sess, err := session1.NewSessionWithOptions(session1.Options{
			SharedConfigState: session1.SharedConfigEnable,
		})

		if err != nil {
			log.Printf("Got an error when trying to initiate the aws session")
			return err
			}
		ssmsvc := ssm1.New(sess, aws1.NewConfig())

		//the username parameter must always be protected by KMS
		withDecryption := true
		param, err := ssmsvc.GetParameter(&ssm1.GetParameterInput{
			Name:           &intent.ParameterStore,
			WithDecryption: &withDecryption,
		})

		if err != nil {
			log.Printf("Got an error when trying to read the parameter (%s), is it protected with KMS?",intent.ParameterStore)
			return err
			}
		
		if Debug == true {
			log.Printf("Accessing the following parameters: %s, %s and %s", intent.ParameterStore, dbparam, hostnameparam)
		}

		password := *param.Parameter.Value

		//The db parameter does not need to be encrypted with KMS
		withDecryptionDb := false
		param2, err := ssmsvc.GetParameter(&ssm1.GetParameterInput{
			Name:           &dbparam,
			WithDecryption: &withDecryptionDb,
		})
		if err != nil {
			log.Printf("Got an error when trying to read the parameter (%s)",dbparam)
			return err
		}
		dbName = *param2.Parameter.Value
		
		//The hostname parameter does not need to be encrypted with KMS
		withDecryptionHostname := false
		param3, err := ssmsvc.GetParameter(&ssm1.GetParameterInput{
			Name:           &hostnameparam,
			WithDecryption: &withDecryptionHostname,
		})
		if err != nil {
			log.Printf("Got an error when trying to read the parameter (%s)",hostnameparam)
			return err
		}
		hostname = *param3.Parameter.Value
		//strings.ContainsAny(intent.QueryType,"parameter")

		db, err := sql.Open(intent.Source, username + ":" + password + "@" + hostname + "/" + dbName)
		if err != nil {
			log.Printf("Got an error when trying to initiate the database connection")
			return err
		}

		defer db.Close()

		if err != nil {
			log.Printf("Failed to connect to the database", err)
			return err
		}
		var query string
		slotMap := request.Intent.Slots
		//If it is a slot here we should use the template for the sql query and the slotname
		if len(intent.SlotName) > 0 {

			if Debug == true {
				log.Printf("This looks like to be a intent file with the slotname %s set, will expect that from alexa and use the parameter (from alexa: %s) in the sql query: %s", intent.SlotName, slotMap[intent.SlotName].Value, intent.Query)
			}

			var tpl bytes.Buffer	
			tmpl, err := template.New("query").Parse(intent.Query)
			if err != nil {
				log.Printf("Got an error when trying to initiate the template, is it correct structured?")
				return err
			}
				//err = tmpl.Execute(&tpl, result2)
			err = tmpl.Execute(&tpl, slotMap)
			if err != nil {
				log.Printf("Got an error during templating")
				return err
			}
	

			if Debug == true {
				log.Printf("INFO: After templating the query it looks like this: %s\n", tpl.String())
			}

			query =tpl.String()

		} else {
			query = intent.Query
		}

		rows, err := db.Query(query)
		if err != nil {
			log.Printf("Failed to run the following query: %s", intent.Query)
			return err
		}

		columnNames, err := rows.Columns()
		if err != nil {
			return err
		}

		rc := NewMapStringScan(columnNames)
		for rows.Next() {
				/*
				cv, err := rowMapString(columnNames, rows)
		      	fck(err)
				if err != nil {
					return err
				}*/

			err := rc.Update(rows)
			if err != nil {
				return err
			}
			//	cv := rc.Get()
		}	
		
		//log.Printf("%#v\n\n" ,rc.Get())
		

		//Starting the templating processing with the result from the database
		var tpl bytes.Buffer	

		tmpl, err := template.New("test").Parse(intent.Say)
		if err != nil {
			log.Printf("Got an error when trying to initiate the template, is it correct structured?")
			return err
		}
//					err = tmpl.Execute(&tpl, result)
		err = tmpl.Execute(&tpl, rc.Get())
		if err != nil {
			log.Printf("Got an error when trying to run template")
			return err
		}


		log.Printf("Output: %s is a %s source, Will say the following: %s\n", request.Intent.Name, intent.Source, tpl.String())
				
		response.SetSimpleCard("Alexa framwork for n00bs github project", tpl.String())
		response.SetOutputText(tpl.String())
		response.SetRepromptText(tpl.String())


	
		

	} else {
		log.Printf("ERROR: Got the following intent that did not match the Souce value (wrong structure or missing fields?): %s with following source: %s \n", request.Intent.Name, intent.Source);
  }
  return nil
}