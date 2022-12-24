package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/naveensrinivasan/rekor-phren/pkg"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// auth provider gcp
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	app := &cli.App{
		Name: "missing-entries",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dataset",
				Usage:   "Name of the dataset",
				Value:   "phren",
				EnvVars: []string{"DATASET"},
			},
			&cli.StringFlag{
				Name:    "table",
				Usage:   "Name of the table",
				Value:   "phren",
				EnvVars: []string{"TABLE"},
			},
		},
		Action: func(c *cli.Context) error {
			dataset := c.String("dataset")
			tableName := c.String("table")

			missing, err := pkg.GetMissingEntries(dataset, tableName)
			if err != nil {
				return err
			}
			if len(missing) == 0 {
				log.Println("No missing entries found")
				return nil
			}
			fmt.Println(missing)
			for i, id := range missing {
				createJob(int(id))
				if i%100 == 0 {
					log.Println("exiting after 100 jobs")
					break
				}
			}
			time.Sleep(5 * time.Second)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
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
func createJob(id int) {
	config, err := buildConfig("")
	if err != nil {
		panic(err.Error())
	}
	deleteJobTime := int32(60)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	jobsClient := clientset.BatchV1().Jobs("default")
	const image = "gcr.io/openssf/rekor-phren-c5fc4a6e85fec69cce84b35fd28b14cc@sha256:4a52ce50e4e240b84d69e04f87d3df684f03a21640cb9229bd4fe8f63b1afc43"
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("job-%d", id),
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &deleteJobTime,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: image,
							Command: []string{"rekor-phren", "--bigquery-dataset", "phren", "--bigquery-table-name",
								"rekor", "--rekor-url", "http://10.117.1.69", "--start-index", fmt.Sprintf("%d", id),
								"--end-index", fmt.Sprintf("%d", id+1), "update"},
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
		panic(err)
	}
	fmt.Printf("Created job %q. \n", result.GetObjectMeta().GetName())
}
