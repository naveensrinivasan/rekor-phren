package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/naveensrinivasan/rekor-phren/pkg"
)

var retry = 5

var e *log.Logger

func main() {
	e = log.New(os.Stdout, "", 0)
	var wg sync.WaitGroup
	var err error
	x := os.Getenv("START")
	y := os.Getenv("END")
	url := os.Getenv("URL")
	tableName := os.Getenv("TABLE_NAME")
	bucketName := os.Getenv("BUCKET_NAME")
	enableRetry := os.Getenv("ENABLE_RETRY")
	if enableRetry != "" {
		retry, err = strconv.Atoi(enableRetry)
		if err != nil {
			e.Println(err)
		}
	}
	if bucketName == "" {
		//nolint
		bucketName = "openssf-rekor-test"
	}
	bucket, err2 := pkg.NewBucket(bucketName)
	if err2 != nil {
		panic(err2)
	}

	if data, ok := os.LookupEnv("RETRY"); ok {
		retry, err = strconv.Atoi(data)
		if err != nil {
			panic(fmt.Errorf("RETRY must be an integer %w", err))
		}
	}
	if tableName == "" {
		//nolint
		tableName = "rekor_test"
	}
	rekor := pkg.NewTLog(url)
	start := int64(0)
	end, err := rekor.Size()
	if err != nil {
		panic(err)
	}
	if x != "" {
		start, err = strconv.ParseInt(x, 10, 64)
		if err != nil {
			panic(err)
		}
	}
	if y != "" {
		end, err = strconv.ParseInt(y, 10, 64)
		if err != nil {
			panic(err)
		}
	}

	for i := start; i <= end; i++ {
		GetRekorEntry(rekor, i, &wg, tableName, bucket)
		time.Sleep(time.Second * 5)
	}
}

// GetRekorEntry gets the rekor entry and updates the table
func GetRekorEntry(rekor pkg.TLog, i int64, wg *sync.WaitGroup, tableName string, bucket pkg.Bucket) {
	data, err := rekor.Entry(i)
	if retry > 0 && err != nil {
		// retrying once more
		time.Sleep(5 * time.Second)
		data, err = rekor.Entry(i)
		if err != nil {
			handleErr(err)
		}
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := pkg.Insert(data, tableName)
		if err != nil {
			handleErr(err)
		}
	}()
	go func() {
		defer wg.Done()
		err := bucket.UpdateBucket(data)
		if err != nil {
			handleErr(err)
		}
	}()
	wg.Wait()
}

// handlerErr handles the error
func handleErr(err error) {
	if err != nil {
		e.Printf("failed to update table %v, skipping", err)
	}
}
