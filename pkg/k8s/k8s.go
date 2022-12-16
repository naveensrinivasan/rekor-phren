package k8s

import (
	"context"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"

	"log"
	"math"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/naveensrinivasan/rekor-phren/pkg"
)

type Range struct {
	From int64
	To   int64
}

// K8s is the interface for the k8s client
type K8s interface {
	// GetPendingRanges returns the ranges for start and end for the given chunkSize which can be parallelized.
	GetPendingRanges(dataset, table string, chunkSize int64) ([]Range, error)
	// CreateJob creates a job in the given namespace within the k8s cluster.
	CreateJob(r Range) error
}
type k8s struct {
	phren     pkg.Phren
	tlog      pkg.TLog
	hostname  string
	dataset   string
	table     string
	namespace string
}

// New returns a new instance of K8s
func New(phren pkg.Phren, tlog pkg.TLog, hostname, dataset, table, namespace string) (K8s, error) {
	// validate the inputs
	if phren == nil {
		return nil, fmt.Errorf("phren cannot be nil")
	}
	if tlog == nil {
		return nil, fmt.Errorf("tlog cannot be nil")
	}
	if hostname == "" {
		return nil, fmt.Errorf("hostname cannot be empty")
	}
	if dataset == "" {
		return nil, fmt.Errorf("dataset cannot be empty")
	}
	if table == "" {
		return nil, fmt.Errorf("table cannot be empty")
	}
	return &k8s{phren: phren, tlog: tlog, hostname: hostname, dataset: dataset, table: table, namespace: namespace}, nil
}

// GetPendingRanges returns the ranges for start and end for the given chunkSize which can be parallelized.
func (k k8s) GetPendingRanges(dataset, table string, chunkSize int64) ([]Range, error) {
	if dataset == "" {
		return nil, fmt.Errorf("dataset cannot be empty")
	}
	if table == "" {
		return nil, fmt.Errorf("table cannot be empty")
	}
	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunkSize cannot be negative")
	}

	result := []Range{}
	// get the max index from the phren table.
	lastEntry, err := k.phren.GetLastEntry(dataset, table)
	log.Printf("last entry: %d\n", lastEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to get last entry: %w", err)
	}
	// get the max index from the tlog.
	tlogSize, err := k.tlog.Size()
	log.Printf("tlog size: %d\n", tlogSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get size of tlog: %w", err)
	}
	lastStart := int64(0)
	// calculate the number of chunks
	for i := (lastEntry/chunkSize)*chunkSize + chunkSize; i < tlogSize-1; i += chunkSize {
		n1 := i - (chunkSize - 1)
		start := int64(math.Max(float64(n1), float64(lastEntry)))
		r := Range{From: start, To: i}
		result = append(result, r)
		lastStart = i + 1
	}
	if len(result) == 0 {
		// there are no pending ranges
		return nil, nil
	}
	result = append(result, Range{From: lastStart, To: tlogSize})
	return result, nil
}

// CreateJob creates a job in the given namespace within the k8s cluster.
func (k k8s) CreateJob(r Range) error {
	config, err := buildConfig("")
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	deleteJobTime := int32(0)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}
	jobsClient := clientset.BatchV1().Jobs("default")
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("phren-%d-%d", r.From, r.To),
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &deleteJobTime,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "phren",
							Image: "gcr.io/openssf/rekor-phren-c5fc4a6e85fec69cce84b35fd28b14cc@sha256:4a52ce50e4e240b84d69e04f87d3df684f03a21640cb9229bd4fe8f63b1afc43",
							Command: []string{"rekor-phren", "--bigquery-dataset", k.dataset, "--bigquery-table-name",
								k.table, "--rekor-url", k.hostname, "--start-index", fmt.Sprintf("%d", r.From), "--end-index", fmt.Sprintf("%d", r.To), "update"},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: "phren",
				},
			},
		},
	}
	result, err := jobsClient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	log.Printf("Created job %q. \n", result.GetObjectMeta().GetName())
	return nil
}
func buildConfig(kubeconfig string) (*rest.Config, error) {
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return rest.InClusterConfig()
	}
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
}
