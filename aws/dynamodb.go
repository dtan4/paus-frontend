package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DynamoDB struct {
	svc *dynamodb.DynamoDB
}

// NewDynamoDB creates new DynamoDB object
func NewDynamoDB() *DynamoDB {
	return &DynamoDB{
		svc: dynamodb.New(session.New(), &aws.Config{}),
	}
}

// Delete deletes the given item
func (d *DynamoDB) Delete(table string, filter map[string]string) error {
	key := make(map[string]*dynamodb.AttributeValue)

	for k, v := range filter {
		key[k] = &dynamodb.AttributeValue{
			S: aws.String(v),
		}
	}

	_, err := d.svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key:       key,
	})
	if err != nil {
		return err
	}

	return nil
}

// List returns all items in the given table
func (d *DynamoDB) List(table string) ([]map[string]*dynamodb.AttributeValue, error) {
	resp, err := d.svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(table),
	})
	if err != nil {
		return []map[string]*dynamodb.AttributeValue{}, err
	}

	return resp.Items, nil
}

// Select returns matched items in the given table
func (d *DynamoDB) Select(table, index string, filter map[string]string) ([]map[string]*dynamodb.AttributeValue, error) {
	keyConditions := make(map[string]*dynamodb.Condition)

	for k, v := range filter {
		keyConditions[k] = &dynamodb.Condition{
			ComparisonOperator: aws.String(dynamodb.ComparisonOperatorEq),
			AttributeValueList: []*dynamodb.AttributeValue{
				&dynamodb.AttributeValue{
					S: aws.String(v),
				},
			},
		}
	}

	params := &dynamodb.QueryInput{
		TableName:     aws.String(table),
		KeyConditions: keyConditions,
	}

	if index != "" {
		params.IndexName = aws.String(index)
	}

	resp, err := d.svc.Query(params)
	if err != nil {
		return []map[string]*dynamodb.AttributeValue{}, err
	}

	return resp.Items, nil
}

// Update updates / creates item in the given table
func (d *DynamoDB) Update(table string, fields map[string]string) error {
	key := make(map[string]*dynamodb.AttributeValue)

	for k, v := range fields {
		key[k] = &dynamodb.AttributeValue{
			S: aws.String(v),
		}
	}

	_, err := d.svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key:       key,
	})
	if err != nil {
		return err
	}

	return nil
}
