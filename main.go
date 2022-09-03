package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/naveensrinivasan/rekor-phren/pkg"
)

var (
	retry = 5
)

func main() {
	var err error
	e := log.New(os.Stderr, "", 0)
	x := os.Getenv("START")
	y := os.Getenv("END")
	url := os.Getenv("URL")
	tableName := os.Getenv("TABLE_NAME")

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
		data, err := rekor.Entry(i)
		if err != nil {
			// retrying once more
			time.Sleep(time.Duration(retry) * time.Second)
			data, err = rekor.Entry(i)
			if err != nil {
				e.Printf("failed to get entry %d: %v, skipping", i, err)
			}
		}
		e := pkg.Insert(data, tableName)
		if e != nil {
			panic(e)
		}
		time.Sleep(time.Second * 5)
	}
}
