package pkg

import (
	"cloud.google.com/go/bigquery"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
)

type Phren interface {
	GetLastEntry(dataset, table string) (int64, error)
}
type phren struct {
}

func New() Phren {
	return &phren{}
}

// CreateOrUpdateSchema creates a new table in BigQuery. The func detects the project ID from the credentials.
func CreateOrUpdateSchema(entry Entry, dataset, table string) error {
	if dataset == "" {
		return fmt.Errorf("dataset is required")
	}
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "openssf")
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
		if err := UpdateTableSchema(entry, dataset, table); err != nil {
			return err
		}
	}
	return nil
}
func UpdateTableSchema(entry Entry, dataset, table string) error {
	if dataset == "" {
		return fmt.Errorf("dataset is required")
	}
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "openssf")
	if err != nil {
		return err
	}
	s, err := bigquery.InferSchema(entry)
	if err != nil {
		return err
	}
	s = s.Relax()
	tableRef := client.Dataset(dataset).Table(table)
	update := bigquery.TableMetadataToUpdate{
		Schema: s,
	}
	if _, err := tableRef.Update(ctx, update, ""); err != nil {
		return fmt.Errorf("tableRef.Update: %w", err)
	}
	return nil
}
func Insert(entry Entry, dataset, table string) error {
	if dataset == "" {
		return fmt.Errorf("dataset is required")
	}
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "openssf")
	if err != nil {
		return err
	}
	inserter := client.Dataset(dataset).Table(table).Inserter()
	if err := inserter.Put(ctx, entry); err != nil {
		return err
	}
	return nil
}

// GetLastEntry returns the last entry from the BigQuery table.
func (p phren) GetLastEntry(dataset, table string) (int64, error) {
	if dataset == "" {
		return 0, fmt.Errorf("dataset is required")
	}
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "openssf")
	if err != nil {
		return 0, fmt.Errorf("bigquery.NewClient: %w", err)
	}
	q := client.Query(fmt.Sprintf("SELECT max(logindex) FROM `openssf.%s.%s` LIMIT 1", dataset, table))
	it, err := q.Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("Query.Read: %w", err)
	}
	var max int64
	for {
		var values []bigquery.Value
		err := it.Next(&values)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("Iterator.Next: %w", err)
		}
		max = values[0].(int64)
	}
	return max, nil
}

// GetMissingEntries returns the missing entries from the BigQuery table.
// This will be used to fill the missing entries in the BigQuery table by rerunning the missing entries as cron job.
func GetMissingEntries(dataset, table string) ([]int64, error) {
	query := "SELECT date FROM UNNEST(ARRAY_CONCAT(" +
		"GENERATE_ARRAY(0, 1000000)" +
		",GENERATE_ARRAY(1000001, 2000000)," +
		"GENERATE_ARRAY(2000001, 3000000)," +
		"GENERATE_ARRAY(3000001, 4000000)," +
		"GENERATE_ARRAY(4000001, 5000000)," +
		"GENERATE_ARRAY(5000001, 6000000)," +
		"GENERATE_ARRAY(6000001, 7000000)," +
		"GENERATE_ARRAY(7000001, 8000000)," +
		"GENERATE_ARRAY(8000001, (select max(logindex) from `openssf.%s.%s`)))) " +
		"date EXCEPT DISTINCT " +
		"SELECT logindex FROM `openssf.%s.%s`;"
	if dataset == "" {
		return nil, fmt.Errorf("dataset is required")
	}
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "openssf")
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %w", err)
	}
	q := client.Query(fmt.Sprintf(query, dataset, table, dataset, table))

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("Query.Read: %w", err)
	}
	var missing []int64
	for {
		var values []bigquery.Value
		err := it.Next(&values)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Iterator.Next: %w", err)
		}
		missing = append(missing, values[0].(int64))
	}
	return missing, nil
}
func getmaxlogindex(dataset, tableName string) (int64, error) {
	max := int64(0)
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "openssf")
	if err != nil {
		return 0, fmt.Errorf("bigquery.NewClient: %w", err)
	}
	q := client.Query(fmt.Sprintf("SELECT max(logindex) FROM `openssf.%s.%s` LIMIT 1", dataset, tableName))
	it, err := q.Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("Query.Read: %w", err)
	}
	for {
		var values []bigquery.Value
		err := it.Next(&values)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("Iterator.Next: %w", err)
		}
		max = values[0].(int64)
	}
	return max, nil
}
