package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBStore struct {
	client    *dynamodb.Client
	tableName string
	ttl       time.Duration
}

func NewDynamoDBStore(client *dynamodb.Client, tableName string, ttl time.Duration) *DynamoDBStore {
	return &DynamoDBStore{client: client, tableName: tableName, ttl: ttl}
}

func (s *DynamoDBStore) Create(targetID string) (*Game, error) {
	ctx := context.Background()
	g := &Game{
		ID:           newDynamoID(),
		TargetID:     targetID,
		LastAccessed: time.Now(),
	}
	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      gameToItem(g, s.ttl),
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *DynamoDBStore) Get(id string) (*Game, error) {
	ctx := context.Background()
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"gameId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}
	g, err := itemToGame(out.Item)
	if err != nil {
		return nil, err
	}
	if time.Since(g.LastAccessed) > s.ttl {
		return nil, ErrNotFound
	}
	g.LastAccessed = time.Now()
	if err := s.Update(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *DynamoDBStore) Update(g *Game) error {
	ctx := context.Background()
	// ConditionExpression ensures the item exists; otherwise return ErrNotFound.
	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                gameToItem(g, s.ttl),
		ConditionExpression: aws.String("attribute_exists(gameId)"),
	})
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func gameToItem(g *Game, ttl time.Duration) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"gameId":       &types.AttributeValueMemberS{Value: g.ID},
		"targetId":     &types.AttributeValueMemberS{Value: g.TargetID},
		"attempts":     &types.AttributeValueMemberN{Value: strconv.Itoa(g.Attempts)},
		"won":          &types.AttributeValueMemberBOOL{Value: g.Won},
		"lastAccessed": &types.AttributeValueMemberN{Value: strconv.FormatInt(g.LastAccessed.Unix(), 10)},
		"expiresAt":    &types.AttributeValueMemberN{Value: strconv.FormatInt(g.LastAccessed.Add(ttl).Unix(), 10)},
	}
}

func itemToGame(item map[string]types.AttributeValue) (*Game, error) {
	g := &Game{}
	if v, ok := item["gameId"].(*types.AttributeValueMemberS); ok {
		g.ID = v.Value
	}
	if v, ok := item["targetId"].(*types.AttributeValueMemberS); ok {
		g.TargetID = v.Value
	}
	if v, ok := item["attempts"].(*types.AttributeValueMemberN); ok {
		n, err := strconv.Atoi(v.Value)
		if err != nil {
			return nil, fmt.Errorf("attempts: %w", err)
		}
		g.Attempts = n
	}
	if v, ok := item["won"].(*types.AttributeValueMemberBOOL); ok {
		g.Won = v.Value
	}
	if v, ok := item["lastAccessed"].(*types.AttributeValueMemberN); ok {
		ts, err := strconv.ParseInt(v.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("lastAccessed: %w", err)
		}
		g.LastAccessed = time.Unix(ts, 0)
	}
	return g, nil
}

func newDynamoID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
