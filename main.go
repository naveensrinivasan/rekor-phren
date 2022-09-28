package main

import (
	"fmt"
	"github.com/naveensrinivasan/rekor-phren/pkg"
	"log"
	"os"
	"strconv"
	"sync"
)

var retry = 5

var e *log.Logger
var tableName string
var bucket pkg.Bucket
var rekor pkg.TLog

func main() {
	e = log.New(os.Stdout, "", 0)
	var err error
	x := os.Getenv("START")
	y := os.Getenv("END")
	z := os.Getenv("CONCURRENCY")

	fmt.Println(x, y)
	url := os.Getenv("URL")
	tableName = os.Getenv("TABLE_NAME")
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
	bucket, err = pkg.NewBucket(bucketName)
	if err != nil {
		panic(err)
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
	rekor = pkg.NewTLog(url)
	start := int64(0)
	end, err := rekor.Size()
	counter := 10
	if z != "" {
		counter, err = strconv.Atoi(z)
		if err != nil {
			panic(err)
		}
	}
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
	// three consumers
	wg := sync.WaitGroup{}
	wg.Add(counter)
	var ch = make(chan int64)

	//consumer
	for i := 0; i < counter; i++ {
		go func() {
			for i := range ch {
				GetRekorEntry(rekor, i, tableName, bucket)
			}
			wg.Done()
		}()
	}

	//producer
	go func() {
		for i := start; i <= end; i++ {
			ch <- i
		}
		close(ch)
	}()

	wg.Wait()

}

// GetRekorEntry gets the rekor entry and updates the table
func GetRekorEntry(rekor pkg.TLog, i int64, tableName string, bucket pkg.Bucket) {
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
	go func(i int64) {
		defer wg.Done()
		err := pkg.Insert(data, tableName)
		if err != nil {
			handleErr(fmt.Errorf("failed to insert entry %d %w", i, err))
		}
	}(i)
	go func(i int64) {
		defer wg.Done()
		err := bucket.UpdateBucket(data)
		if err != nil {
			handleErr(fmt.Errorf("failed to update bucket %d %w", i, err))
		}
	}(i)
	if i%100 == 0 {
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
