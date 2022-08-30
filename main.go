package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/naveensrinivasan/rekor-phren/pkg"
)

func main() {
	x := os.Getenv("START")
	y := os.Getenv("END")
	url := os.Getenv("URL")
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
	for i := start; i < end; i++ {
		data, err := rekor.Entry(i)
		if err != nil {
			fmt.Println(err)
		}
		serialized, err := Marshal(data)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(serialized))
		time.Sleep(time.Second * 5)
	}
}

// Marshal is a UTF-8 friendly marshaller.  Go's json.Marshal is not UTF-8
// friendly because it replaces the valid UTF-8 and JSON characters "&". "<",
// ">" with the "slash u" unicode escaped forms (e.g. \u0026).  It preemptively
// escapes for HTML friendliness.  Where text may include any of these
// characters, json.Marshal should not be used. Playground of Go breaking a
// title: https://play.golang.org/p/o2hiX0c62oN
func Marshal(i interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(i)
	return bytes.TrimRight(buffer.Bytes(), "\n"), err
}
