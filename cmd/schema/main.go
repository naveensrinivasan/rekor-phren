package main

import (
	"github.com/naveensrinivasan/rekor-phren/pkg"
	"os"
)

func main() {
	var datasetName, tableName string
	datasetName = "phren"
	tableName = "rekor"
	if len(os.Args) > 3 {
		datasetName = os.Args[1]
		tableName = os.Args[2]
	}
	// this func updates the BigQuery table schema
	k := pkg.Entry{}
	err := pkg.CreateOrUpdateSchema(k, datasetName, tableName)
	if err != nil {
		panic(err)
	}
}
