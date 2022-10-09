package main

import (
	"fmt"
	"github.com/naveensrinivasan/rekor-phren/pkg"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sync"
)

var (
	retry      = 5
	e          *log.Logger
	tableName  = "rekor_test"
	bucket     pkg.Bucket
	rekor      pkg.TLog
	url        string
	bucketName = "openssf-rekor-test"
)

func main() {
	start, end, concurrency := 0, 0, 10
	e = log.New(os.Stdout, "", 0)

	app := &cli.App{
		Name:  "rekor-phren is a tool to update the BigQuery table and the bucket with the rekor entries",
		Usage: "rekor-phren update -u <rekor url> -b <bucket name> -t <table name> -s <start> -e <end> -c <concurrency>",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "start-index",
				DefaultText: "0",
				Aliases:     []string{"s"},
				Value:       0,
				Destination: &start,
				EnvVars: []string{
					"PHREN_START",
				},
			},
			&cli.IntFlag{
				Name:        "end-index",
				Aliases:     []string{"e"},
				Destination: &end,
				EnvVars: []string{
					"PHREN_END",
				},
			},
			&cli.IntFlag{
				Name:        "concurrency-level-update",
				Aliases:     []string{"c"},
				Value:       concurrency,
				Destination: &concurrency,
				DefaultText: "10",
				EnvVars: []string{
					"PHREN_CONCURRENCY",
				},
			},
			&cli.StringFlag{
				Name:        "rekor-url",
				Aliases:     []string{"u"},
				Value:       url,
				DefaultText: "https://api.rekor.sigstore.dev",
				Destination: &url,
				EnvVars: []string{
					"REKOR_URL",
				},
			},
			&cli.StringFlag{
				Name:        "bigquery-table-name",
				Aliases:     []string{"t"},
				Value:       tableName,
				Destination: &tableName,
				EnvVars: []string{
					"PHREN_TABLE",
				},
			},
			&cli.StringFlag{
				Name:        "gcs-bucket-name",
				Aliases:     []string{"b"},
				DefaultText: "openssf-rekor-test",
				Value:       bucketName,
				Destination: &bucketName,
				EnvVars: []string{
					"PHREN_BUCKET",
				},
			},
			&cli.IntFlag{
				Name:        "number-of-retries",
				Aliases:     []string{"r"},
				Value:       retry,
				DefaultText: "5",
				Destination: &retry,
				EnvVars: []string{
					"PHREN_RETRY",
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "update",
				Usage:       "update -u <rekor url> -b <bucket name> -t <table name> -s <start> -e <end> -c <concurrency>",
				Description: "This command updates the BigQuery table and the bucket with the rekor entries. ",
				Action: func(c *cli.Context) error {
					update(end, concurrency, start)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func update(end int, concurrency int, start int) error {
	var err error
	if bucketName == "" {
		bucketName = "openssf-rekor-test"
	}
	rekor = pkg.NewTLog(url)
	if end == 0 {
		end, err = rekor.Size()
		if err != nil {
			return fmt.Errorf("failed to get the size of the rekor log %w", err)
		}
	}

	bucket, err := pkg.NewBucket(bucketName)
	if err != nil {
		return fmt.Errorf("failed to create bucket %w", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	var ch = make(chan int)

	// consumer
	for i := 0; i < concurrency; i++ {
		go func() {
			for i := range ch {
				GetRekorEntry(rekor, i, tableName, bucket)
			}
			wg.Done()
		}()
	}

	// producer
	go func() {
		for i := start; i <= end; i++ {
			ch <- i
		}
		close(ch)
	}()

	wg.Wait()
	return nil
}

// GetRekorEntry gets the rekor entry and updates the table
func GetRekorEntry(rekor pkg.TLog, i int, tableName string, bucket pkg.Bucket) {
	var wg sync.WaitGroup
	data, err := rekor.Entry(i)
	if retry > 0 && err != nil {
		// retrying once more
		data, err = rekor.Entry(i)
		if err != nil {
			handleErr(err)
		}
	}
	wg.Add(2)
	go func(i int) {
		defer wg.Done()
		err := pkg.Insert(data, tableName)
		if err != nil {
			handleErr(fmt.Errorf("failed to insert entry %d %w", i, err))
		}
	}(i)
	go func(i int) {
		defer wg.Done()
		err := bucket.UpdateBucket(data)
		if err != nil {
			handleErr(fmt.Errorf("failed to update bucket %d %w", i, err))
		}
	}(i)
	if i%1000 == 0 {
		fmt.Println("Finished", i)
	}
	wg.Wait()
}

// handlerErr handles the error
func handleErr(err error) {
	if err != nil {
		e.Printf("failed to update table %v, skipping", err)
	}
}
