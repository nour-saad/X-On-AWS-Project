package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var svc *dynamodb.DynamoDB

type Tweet struct {
	ID        string    `json:"id" dynamodbav:"ID"`
	UserID    string    `json:"user_id" dynamodbav:"UserID"`
	Content   string    `json:"content" dynamodbav:"Content"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"CreatedAt"`
}

func main() {
	connectDynamo()
	createTable()

	r := mux.NewRouter()
	r.HandleFunc("/tweets", createTweet).Methods("POST")
	r.HandleFunc("/tweets/{userID}", getTweetsByUser).Methods("GET")

	http.ListenAndServe(":8081", r)
}

func connectDynamo() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Endpoint:    aws.String(os.Getenv("DYNAMODB_ENDPOINT")),
		Credentials: credentials.NewStaticCredentials("dummy", "dummy", ""),
	}))

	svc = dynamodb.New(sess)
}

func createTable() {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String("tweets"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("ID"), AttributeType: aws.String("S")},
			{AttributeName: aws.String("UserID"), AttributeType: aws.String("S")},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("ID"), KeyType: aws.String("HASH")},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("UserIDIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{AttributeName: aws.String("UserID"), KeyType: aws.String("HASH")},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String("ALL"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := svc.CreateTable(input)
	if err != nil && !isTableExistsError(err) {
		panic(err.Error())
	}
}

func isTableExistsError(err error) bool {
	var tae *dynamodb.ResourceInUseException
	return errors.As(err, &tae)
}

func createTweet(w http.ResponseWriter, r *http.Request) {
	var tweet Tweet
	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate unique ID and set timestamp
	tweet.ID = uuid.New().String()
	tweet.CreatedAt = time.Now()

	// Marshal to DynamoDB format
	item, err := dynamodbattribute.MarshalMap(tweet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String("tweets"),
		Item:      item,
	}

	// Execute PutItem
	_, err = svc.PutItem(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tweet)
}

func getTweetsByUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	result, err := svc.Query(&dynamodb.QueryInput{
		TableName:              aws.String("tweets"),
		IndexName:              aws.String("UserIDIndex"),
		KeyConditionExpression: aws.String("UserID = :uid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {S: aws.String(userID)},
		},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tweets []Tweet
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &tweets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tweets)
}

// Add HTTP handlers and helper functions here
