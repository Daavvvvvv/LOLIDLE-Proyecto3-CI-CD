package lore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBCache struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBCache(client *dynamodb.Client, tableName string) *DynamoDBCache {
	return &DynamoDBCache{client: client, tableName: tableName}
}

func (c *DynamoDBCache) Get(ctx context.Context, championID string) (string, bool, error) {
	out, err := c.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.tableName),
		Key: map[string]types.AttributeValue{
			"championId": &types.AttributeValueMemberS{Value: championID},
		},
	})
	if err != nil {
		return "", false, err
	}
	if out.Item == nil {
		return "", false, nil
	}
	v, ok := out.Item["lore"].(*types.AttributeValueMemberS)
	if !ok {
		return "", false, nil
	}
	return v.Value, true, nil
}

func (c *DynamoDBCache) Put(ctx context.Context, championID, lore string) error {
	_, err := c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.tableName),
		Item: map[string]types.AttributeValue{
			"championId": &types.AttributeValueMemberS{Value: championID},
			"lore":       &types.AttributeValueMemberS{Value: lore},
		},
	})
	return err
}
