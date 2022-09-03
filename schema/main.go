package main

import (
	"github.com/naveensrinivasan/rekor-phren/pkg"
)

//nolint:funlen
func main() {
	// this func updates the BigQuery table schema
	k := pkg.Entry{}
	err := pkg.CreateOrUpdateSchema(k, "rekor_test")
	if err != nil {
		panic(err)
	}
}
