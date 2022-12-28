package dynamo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
)

var _ LeaderboardRepo = (*LeaderboardDB)(nil)

type LeaderboardDB struct {
	client    *dynamodb.Client
	tableName string
}

type LeaderboardEntry struct {
	TypeName  string  `dynamodbav:"__typename"`
	UserID    string  `dynamodbav:"userId"`
	Username  string  `dynamodbav:"username"`
	Count     int     `dynamodbav:"count"`
	Volume    float64 `dynamodbav:"volume"`
	UpdatedAt string  `dynamodbav:"updatedAt"`
}

func NewLeaderboardRepo(client *dynamodb.Client, tableName string) *LeaderboardDB {
	return &LeaderboardDB{
		client:    client,
		tableName: tableName,
	}
}

func (r *LeaderboardDB) Get(ctx context.Context, userID string) (*LeaderboardEntry, error) {
	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: userID},
		},
	}

	getItemOutput, err := r.client.GetItem(ctx, getItemInput)
	if err != nil {
		return nil, errors.Wrap(err, "could not get leaderboard entry item")
	}

	if getItemOutput.Item == nil || len(getItemOutput.Item) == 0 {
		return nil, nil
	}

	leaderboardEntry := &LeaderboardEntry{}

	err = attributevalue.UnmarshalMap(getItemOutput.Item, leaderboardEntry)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal leaderboard entry item")
	}

	return leaderboardEntry, nil
}

func (r *LeaderboardDB) GetAll(ctx context.Context) ([]LeaderboardEntry, error) {
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	scanOutput, err := r.client.Scan(ctx, scanInput)
	if err != nil {
		return nil, errors.Wrap(err, "could not scan leaderboard entry items")
	}

	if scanOutput == nil || scanOutput.Items == nil || len(scanOutput.Items) == 0 {
		return nil, nil
	}

	leaderboardEntries := []LeaderboardEntry{}

	err = attributevalue.UnmarshalListOfMaps(scanOutput.Items, &leaderboardEntries)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal leaderboard entry items")
	}

	return leaderboardEntries, nil
}

func (r *LeaderboardDB) Save(ctx context.Context, leaderboardEntry *LeaderboardEntry) error {
	leaderboardEntry.TypeName = "LeaderboardEntry"

	if leaderboardEntry.UserID == "" {
		return errors.New("userId is required")
	}

	if leaderboardEntry.Username == "" {
		return errors.New("username is required")
	}

	leaderboardEntry.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	avMap, err := attributevalue.MarshalMap(leaderboardEntry)
	if err != nil {
		return errors.Wrap(err, "could not marshal leaderboard entry item")
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      avMap,
	}

	if _, err := r.client.PutItem(ctx, putItemInput); err != nil {
		return errors.Wrap(err, "could put leaderboard entry item")
	}

	return nil
}
