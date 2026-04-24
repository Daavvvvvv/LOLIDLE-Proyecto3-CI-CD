package session

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func newDynamoTestClient(t *testing.T) *dynamodb.Client {
	t.Helper()
	endpoint := os.Getenv("DYNAMO_LOCAL_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
	)
	if err != nil {
		t.Skipf("cannot create AWS config: %v", err)
	}
	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
}

func setupTable(t *testing.T, client *dynamodb.Client, name string) {
	t.Helper()
	_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String(name)})
	_, err := client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName: aws.String(name),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("gameId"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("gameId"), KeyType: types.KeyTypeHash},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("CreateTable: %v", err)
	}
	t.Cleanup(func() {
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{TableName: aws.String(name)})
	})
}

func TestDynamoDBStore_Create_storesAndReturnsGame(t *testing.T) {
	client := newDynamoTestClient(t)
	setupTable(t, client, "test-sessions-1")
	store := NewDynamoDBStore(client, "test-sessions-1", time.Minute)

	g, err := store.Create("ahri")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if g.ID == "" {
		t.Error("expected non-empty ID")
	}
	if g.TargetID != "ahri" {
		t.Errorf("TargetID = %s, want ahri", g.TargetID)
	}
}

func TestDynamoDBStore_Get_returnsCreatedGame(t *testing.T) {
	client := newDynamoTestClient(t)
	setupTable(t, client, "test-sessions-2")
	store := NewDynamoDBStore(client, "test-sessions-2", time.Minute)

	created, _ := store.Create("yasuo")
	got, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.TargetID != "yasuo" {
		t.Errorf("TargetID = %s, want yasuo", got.TargetID)
	}
}

func TestDynamoDBStore_Get_returnsErrNotFoundForUnknownID(t *testing.T) {
	client := newDynamoTestClient(t)
	setupTable(t, client, "test-sessions-3")
	store := NewDynamoDBStore(client, "test-sessions-3", time.Minute)

	_, err := store.Get("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDynamoDBStore_Update_persistsChanges(t *testing.T) {
	client := newDynamoTestClient(t)
	setupTable(t, client, "test-sessions-4")
	store := NewDynamoDBStore(client, "test-sessions-4", time.Minute)

	g, _ := store.Create("ahri")
	g.Attempts = 5
	g.Won = true
	if err := store.Update(g); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := store.Get(g.ID)
	if got.Attempts != 5 {
		t.Errorf("Attempts = %d, want 5", got.Attempts)
	}
	if !got.Won {
		t.Error("expected Won=true")
	}
}
