// This package contains the logic to update the BigQuery table and the bucket with the rekor entries.
// This package gets difference between the rekor entries and the BigQuery table.
// It then creates a k8s job to update the BigQuery table and the bucket with the rekor entries.
// It slices the difference into chunks of 50000 entries and creates a k8s job for each chunk.
package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/naveensrinivasan/rekor-phren/pkg"
	"github.com/naveensrinivasan/rekor-phren/pkg/k8s"

	// auth provider gcp
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const chunkSize = 50000
const job = "phren"

func main() {
	// the hostname of the rekor server
	hostname := "http://rekor-sigstore-server.sigstore.svc.cluster.local"
	// the table to use for bigquery
	table := "rekor"
	// the dataset to use for bigquery
	dataset := "phren"
	// the namespace to use for k8s for the job
	namespace := "default"

	app := handleCommandline(hostname, dataset, table, namespace)
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// handleCommandline handles the commandline arguments
//nolint:funlen
func handleCommandline(hostname string, dataset string, table string, namespace string) *cli.App {
	app := cli.NewApp()
	app.Name = "phren-scan"
	app.Usage = "phren scan looks for new entries in rekor and invokes the phren job to update the BigQuery table and the bucket with the rekor entries"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "rekor URL",
			Value:       hostname,
			Usage:       "URL to the rekor server",
			Destination: &hostname,
		},
		&cli.StringFlag{
			Name:        "dataset",
			Value:       dataset,
			Usage:       "dataset to use for bigquery",
			Destination: &dataset,
		},
		&cli.StringFlag{
			Name:        "table",
			Value:       table,
			Usage:       "table to use table for bigquery",
			Destination: &table,
		},
		&cli.StringFlag{
			Name:        "namespace",
			Value:       namespace,
			Usage:       "namespace to use for k8s",
			Destination: &namespace,
		},
	}

	app.Action = func(c *cli.Context) error {
		// gets all the k8s jobs that are running
		jobs, err := RunningK8sJobs()
		if err != nil {
			return fmt.Errorf("error getting running k8s jobs: %w", err)
		}
		// If there are any jobs running, exit.
		for _, j := range jobs {
			if strings.Contains(j, job) {
				log.Printf("job %s is already running\n so we aren't going to run another job.", j)
				return nil
			}
		}

		p := pkg.New()
		t := pkg.NewTLog(hostname)
		k, err := k8s.New(p, t, hostname, dataset, table, namespace)

		if err != nil {
			return fmt.Errorf("error creating k8s client: %w", err)
		}
		// gets the difference between the rekor entries and the BigQuery table and slices it into chunks of 50000 entries
		result, err := k.GetPendingRanges(dataset, table, chunkSize)

		if err != nil {
			return fmt.Errorf("error getting pending ranges: %w", err)
		}

		if len(result) == 0 {
			log.Println("no pending entries")
			return nil
		}

		for _, r := range result {
			// creates a k8s job for each chunk
			if err := k.CreateJob(r); err != nil {
				panic(err)
			}
		}
		return nil
	}
	return app
}

// RunningK8sJobs returns the list of running k8s jobs
func RunningK8sJobs() ([]string, error) {
	config, err := BuildConfig("")
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	jobs, err := client.BatchV1().Jobs("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var jobNames []string //nolint:prealloc
	for _, job := range jobs.Items {
		jobNames = append(jobNames, job.Name)
	}
	return jobNames, nil
}

// BuildConfig builds the k8s config
func BuildConfig(kubeconfig string) (*rest.Config, error) {
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return rest.InClusterConfig()
	}
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
}
