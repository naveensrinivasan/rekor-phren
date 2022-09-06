package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
)

type Bucket interface {
	UpdateBucket(item Entry) error
}
type bucket struct {
	Name string
}

func NewBucket(name string) (Bucket, error) {
	if name == "" {
		return nil, fmt.Errorf("bucket name is required")
	}
	return &bucket{
		Name: name,
	}, nil
}

func (b bucket) UpdateBucket(item Entry) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	path := fmt.Sprintf("%d/entry.json", item.LogIndex)
	wc := client.Bucket(b.Name).Object(path).NewWriter(ctx)
	wc.ContentType = "application/json"
	json, err := Marshal(item)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}
	if _, err := wc.Write(json); err != nil {
		return fmt.Errorf("Object(%q).Writer: %w", path, err)
	}
	return wc.Close()
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
