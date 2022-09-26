package pkg

import (
	"cloud.google.com/go/bigquery"
	"context"
	"fmt"
)

// CreateOrUpdateSchema creates a new table in BigQuery. The func detects the project ID from the credentials.
func CreateOrUpdateSchema(entry Entry, dataset string) error {
	if dataset == "" {
		return fmt.Errorf("dataset is required")
	}
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, bigquery.DetectProjectID)
	if err != nil {
		return err
	}
	r := client.Dataset(dataset)
	s, err := bigquery.InferSchema(entry)
	if err != nil {
		return err
	}
	s = s.Relax()
	tables := r.Tables(context.Background())
	isTableExists := false
	for {
		t, err := tables.Next()
		if err != nil {
			break
		}
		if t.TableID == dataset {
			isTableExists = true
			break
		}
	}
	if !isTableExists {
		if err := r.Table(dataset).Create(context.Background(),
			&bigquery.TableMetadata{Schema: s}); err != nil {
			return err
		}
	} else {
		if err := UpdateTableSchema(entry, dataset); err != nil {
			return err
		}
	}
	return nil
}
func UpdateTableSchema(entry Entry, dataset string) error {
	if dataset == "" {
		return fmt.Errorf("dataset is required")
	}
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, bigquery.DetectProjectID)
	if err != nil {
		return err
	}
	s, err := bigquery.InferSchema(entry)
	if err != nil {
		return err
	}
	s = s.Relax()
	tableRef := client.Dataset(dataset).Table(dataset)
	update := bigquery.TableMetadataToUpdate{
		Schema: s,
	}
	if _, err := tableRef.Update(ctx, update, ""); err != nil {
		return fmt.Errorf("tableRef.Update: %w", err)
	}
	return nil
}
func Insert(entry Entry, dataset string) error {
	if dataset == "" {
		return fmt.Errorf("dataset is required")
	}
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, bigquery.DetectProjectID)
	if err != nil {
		return err
	}
	inserter := client.Dataset(dataset).Table(dataset).Inserter()
	if err := inserter.Put(ctx, entry); err != nil {
		return err
	}
	return nil
}
