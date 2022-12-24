package dynamo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pkg/errors"
)

var _ BrewRepo = (*BrewDB)(nil)

var cutoff = time.Date(2022, time.December, 15, 0, 0, 0, 0, time.UTC)

type BrewDB struct {
	client    *dynamodb.Client
	tableName string
}

type Brew struct {
	TypeName  string  `dynamodbav:"__typename"`
	ID        string  `dynamodbav:"id"`
	UserID    string  `dynamodbav:"userId"`
	Username  string  `dynamodbav:"username"`
	Style     string  `dynamodbav:"style"`
	Amount    float64 `dynamodbav:"amount"`
	CreatedAt string  `dynamodbav:"createdAt"`
}

func NewBrewRepo(client *dynamodb.Client, tableName string) *BrewDB {
	return &BrewDB{
		client:    client,
		tableName: tableName,
	}
}

func (r *BrewDB) Get(ctx context.Context, id string) (*Brew, error) {
	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	getItemOutput, err := r.client.GetItem(ctx, getItemInput)
	if err != nil {
		return nil, errors.Wrap(err, "could not get item")
	}

	if getItemOutput.Item == nil || len(getItemOutput.Item) == 0 {
		return nil, nil
	}

	brew := &Brew{}

	err = attributevalue.UnmarshalMap(getItemOutput.Item, brew)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal item")
	}

	return brew, nil
}

func (r *BrewDB) GetAll(ctx context.Context) ([]Brew, error) { // TODO: this is inefficient, fix later
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	scanOutput, err := r.client.Scan(ctx, scanInput)
	if err != nil {
		return nil, errors.Wrap(err, "could not scan items")
	}

	if scanOutput == nil || scanOutput.Items == nil || len(scanOutput.Items) == 0 {
		return nil, nil
	}

	brews := []Brew{}

	err = attributevalue.UnmarshalListOfMaps(scanOutput.Items, &brews)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal items")
	}

	result := []Brew{}

	for _, brew := range brews {
		createdAt, err := time.Parse(time.RFC3339, brew.CreatedAt)
		if err != nil {
			return nil, errors.Wrapf(err, "could parse date %s", brew.CreatedAt)
		}

		if createdAt.After(cutoff) {
			result = append(result, brew)
		}
	}

	return result, nil
}

func (r *BrewDB) GetByUserID(ctx context.Context, userID string) ([]Brew, error) {
	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(r.tableName),
		IndexName: aws.String("byUserId"),
		KeyConditions: map[string]types.Condition{
			"userId": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: userID},
				},
			},
		},
	}

	queryOutput, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, errors.Wrap(err, "could not query items")
	}

	if queryOutput == nil || queryOutput.Items == nil || len(queryOutput.Items) == 0 {
		return nil, nil
	}

	brews := []Brew{}

	err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &brews)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal items")
	}

	result := []Brew{}

	for _, brew := range brews {
		createdAt, err := time.Parse(time.RFC3339, brew.CreatedAt)
		if err != nil {
			return nil, errors.Wrapf(err, "could parse date %s", brew.CreatedAt)
		}

		if createdAt.After(cutoff) {
			result = append(result, brew)
		}
	}

	return result, nil
}

func (r *BrewDB) Save(ctx context.Context, brew *Brew) error {
	brew.TypeName = "Brew"

	if brew.ID == "" {
		id, err := gonanoid.New()
		if err != nil {
			return errors.Wrap(err, "could not create uuid")
		}

		brew.ID = id
		brew.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	avMap, err := attributevalue.MarshalMap(brew)
	if err != nil {
		return errors.Wrap(err, "could not marshal item")
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      avMap,
	}

	if _, err := r.client.PutItem(ctx, putItemInput); err != nil {
		return errors.Wrap(err, "could put item")
	}

	return nil
}

func (r *BrewDB) Delete(ctx context.Context, id string) error {
	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	if _, err := r.client.DeleteItem(ctx, deleteItemInput); err != nil {
		return errors.Wrap(err, "could delete item")
	}

	return nil
}
