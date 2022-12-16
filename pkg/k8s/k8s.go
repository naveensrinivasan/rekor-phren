package pkg

import "fmt"

type Range struct {
	From int
	To   int
}

type K8s interface {
	GetPendingRanges(dataset, table string) ([]Range, error)
}
type k8s struct {
	phren,
	tlog,
}

func (k k8s) GetPendingRanges(dataset, table string) ([]Range, error) {
	entry, err := New().GetLastEntry(dataset, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get last entry: %w", err)
	}

	return nil, nil
}
